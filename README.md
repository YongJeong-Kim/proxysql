# proxysql

```bash 
$ docker compose cp ./master/my.cnf master:/etc
$ docker compose cp ./slave/my.cnf slave1:/etc
$ docker compose cp ./slave/my2.cnf slave2:/etc/my.cnf
$ docker compose cp ./slave/my3.cnf slave3:/etc/my.cnf
```

```bash 
# all slaves
mysql> change master to MASTER_HOST='master', MASTER_USER='testuser', MASTER_PASSWORD='1234', MASTER_LOG_FILE='mysql-bin.000003', MASTER_LOG_POS=1153;
```
포트 번호를 별도로 설정하려면 다음과 같은 옵션을 사용한다.
`MASTER_PORT=<PORT_NUM>`

```bash 
mysql> start slave;
```

```bash 
mysql> show slave status\G
```

connect proxysql admin
```bash
$ docker exec -it proxy bash
$ mysql -u admin -P 6032
```
OR
```bash 
$ docker compose exec -it proxy mysql -u admin -p admin -P 6032
```

show server list
```bash 
# proxysql
MySQL [admin]> SELECT * FROM mysql_servers;
Empty set (0.000 sec)
```

add server 
```bash 
# proxysql
MySQL [admin]> INSERT INTO mysql_servers(hostgroup_id, hostname) VALUES (0, 'master');
MySQL [admin]> INSERT INTO mysql_servers(hostgroup_id, hostname) VALUES (1, 'slave1');
MySQL [admin]> INSERT INTO mysql_servers(hostgroup_id, hostname) VALUES (1, 'slave2'); 
```

show server list
```bash 
MySQL [admin]> SELECT * FROM mysql_servers;
+--------------+----------+------+-----------+--------+--------+-------------+-----------------+---------------------+---------+----------------+---------+
| hostgroup_id | hostname | port | gtid_port | status | weight | compression | max_connections | max_replication_lag | use_ssl | max_latency_ms | comment |
+--------------+----------+------+-----------+--------+--------+-------------+-----------------+---------------------+---------+----------------+---------+
| 0            | master   | 3306 | 0         | ONLINE | 1      | 0           | 1000            | 0                   | 0       | 0              |         |
| 1            | slave1   | 3306 | 0         | ONLINE | 1      | 0           | 1000            | 0                   | 0       | 0              |         |
| 1            | slave2   | 3306 | 0         | ONLINE | 1      | 0           | 1000            | 0                   | 0       | 0              |         |
+--------------+----------+------+-----------+--------+--------+-------------+-----------------+---------------------+---------+----------------+---------+
3 rows in set (0.000 sec)
```

변경내용 적용하기
```bash 
# 메모리에 있는 변경내역 불러오기
MySQL [admin]> LOAD MYSQL SERVERS TO RUNTIME;

# 불러온 변경내역 저장
MySQL [admin]> SAVE MYSQL SERVERS TO DISK;
```
```bash
MySQL [admin]> LOAD MYSQL SERVERS TO RUNTIME;
Query OK, 0 rows affected (0.002 sec)

MySQL [admin]> SAVE MYSQL SERVERS TO DISK;
Query OK, 0 rows affected (0.030 sec)
```

proxysql.cnf에 설정한 `monitor_username="monitor"`, `monitor_password="monitor"`
모니터링 계정을 생성하여 서버 상태를 체크할 수 있다.
```bash 
# master
mysql> create user 'monitor'@'%' identified by 'monitor';
```

general log 설정
```bash 
# master 
mysql> SHOW VARIABLES LIKE '%general%';
+------------------+---------------------------------+
| Variable_name    | Value                           |
+------------------+---------------------------------+
| general_log      | OFF                             |
| general_log_file | /var/lib/mysql/f83b85cbf27c.log |
+------------------+---------------------------------+
2 rows in set (0.00 sec)

mysql> SHOW VARIABLES LIKE '%log_output%';
+---------------+-------+
| Variable_name | Value |
+---------------+-------+
| log_output    | FILE  |
+---------------+-------+
1 row in set (0.00 sec)
```
general log는 꺼져있으며 로그는 file로 저장되도록 설정돼있다.
general log를 켜고 file 대신 table로 저장되도록 수정한다
```bash 
# master
mysql> set global general_log = OFF;

# general_log가 mysql DB의 general_log TABLE에 log 기록
mysql> set global log_output = 'TABLE';

# general_log가 /rdsdbdata/log/general/mysql-general.log 위치에 FILE로 기록
mysql> set global log_output = 'FILE';

# general_log가 TABLE과 FILE에 둘다 기록
mysql> set global log_output = 'TABLE,FILE';

# general_log 기능을 활성화
mysql> set global general_log = ON;
```

master에서 만든 monitor 계정이 proxysql의 서버 상태를 확인하는 것을 볼 수 있다 
```bash 
# master 
mysql> select * from mysql.general_log;
+----------------------------+----------------------------------+-----------+-----------+--------------+--------------------------------------------------------------------------+
| event_time                 | user_host                        | thread_id | server_id | command_type | argument                                                                 |
+----------------------------+----------------------------------+-----------+-----------+--------------+--------------------------------------------------------------------------+
| 2024-07-01 14:46:31.987603 | root[root] @ localhost []        |       137 |         1 | Query        | 0x73686F7720646174616261736573                                           |
| 2024-07-01 14:46:36.790895 | root[root] @ localhost []        |       137 |         1 | Query        | 0x73656C656374202A2066726F6D206D7973716C2E67656E6572616C5F6C6F67         |
| 2024-07-01 14:46:42.809148 | [monitor] @  [172.22.0.2]        |       142 |         1 | Connect      | 0x6D6F6E69746F72403137322E32322E302E32206F6E20207573696E67205443502F4950 |
| 2024-07-01 14:46:42.809368 | monitor[monitor] @  [172.22.0.2] |       142 |         1 | Quit         | 0x                                                                       |
| 2024-07-01 14:46:49.970452 | root[root] @ localhost []        |       137 |         1 | Query        | 0x73656C656374202A2066726F6D206D7973716C2E67656E6572616C5F6C6F67         |
+----------------------------+----------------------------------+-----------+-----------+--------------+--------------------------------------------------------------------------+
5 rows in set (0.00 sec)
```

클라이언트(웹서버 등)에서 proxysql로 접속하기 위한 유저 등록하기 
```bash 
# master 
mysql> SELECT host, user, authentication_string FROM mysql.user;
+-----------+------------------+------------------------------------------------------------------------+
| host      | user             | authentication_string                                                  |
+-----------+------------------+------------------------------------------------------------------------+
| %         | monitor          | $A$005$+1[cGAq0\GY
                                                   W)5JIQDFF8UWYIutAbW/BGEbYahXhJdnnyg0tcb3sEJqNoL15 |
| %         | root             | $A$005$u_\VVSOls<r>1<m{wZu6XyMZ7Acqq9Rs910foFplP1Zxiyf9.R0GIghHAh7 |
| %         | testuser         | *A4B6157319038724E3560894F7F932C8886EBFCF                              |
| localhost | mysql.infoschema | $A$005$THISISACOMBINATIONOFINVALIDSALTANDPASSWORDTHATMUSTNEVERBRBEUSED |
| localhost | mysql.session    | $A$005$THISISACOMBINATIONOFINVALIDSALTANDPASSWORDTHATMUSTNEVERBRBEUSED |
| localhost | mysql.sys        | $A$005$THISISACOMBINATIONOFINVALIDSALTANDPASSWORDTHATMUSTNEVERBRBEUSED |
| localhost | root             | $A$005$]CSHZ:jvj%Kkd␦lA36THm8QrXnrcluGzeRpmHuxsmq1nDIkO3vCwGkWm4lPB |
+-----------+------------------+------------------------------------------------------------------------+
7 rows in set (0.00 sec)
```

만들었던 testuser를 등록한다.
`Warning! 새로운 유저를 만드는 것이 좋다`
```bash 
# proxysql
MySQL [admin]> SELECT * FROM mysql_users;
Empty set (0.000 sec)

MySQL [admin]> INSERT INTO mysql_users(username, password) VALUES ('testuser', '*A4B6157319038724E3560894F7F932C8886EBFCF');
Query OK, 1 row affected (0.000 sec)

MySQL [admin]> LOAD MYSQL SERVERS TO RUNTIME;
Query OK, 0 rows affected (0.002 sec)

MySQL [admin]> SAVE MYSQL SERVERS TO DISK;
Query OK, 0 rows affected (0.030 sec)
```

connect client
```bash
# master
$ docker compose exec -it master mysql -u testuser -p -P 16033 -h localhost
```

### query rules
#### 현재 query rule 확인하기 
```bash 
# proxy 
MySQL [admin]> SELECT * FROM mysql_query_rules;
Empty set (0.000 sec)
```

#### query rule 등록하기
위에서 master는 hostgroup_id를 0로, slave는 hostgroup_id를 1로 설정했다
```bash 
# proxy
MySQL [admin]> INSERT INTO mysql_query_rules(match_pattern,destination_hostgroup,active) VALUES ('^INSERT',0,1);
MySQL [admin]> INSERT INTO mysql_query_rules(match_pattern,destination_hostgroup,active) VALUES ('^UPDATE',0,1);
MySQL [admin]> INSERT INTO mysql_query_rules(match_pattern,destination_hostgroup,active) VALUES ('^DELETE',0,1);
MySQL [admin]> INSERT INTO mysql_query_rules(match_pattern,destination_hostgroup,active) VALUES ('^SELECT',1,1);

MySQL [admin]> LOAD MYSQL SERVERS TO RUNTIME;
Query OK, 0 rows affected (0.002 sec)

MySQL [admin]> SAVE MYSQL SERVERS TO DISK;
Query OK, 0 rows affected (0.030 sec)
```

master general log
```bash 

```

slave general log
```bash 

```
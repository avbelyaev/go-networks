## Цель работы
Целью данной работы является разработка одноранговой сетевой службы.

## Задание
Основные требования к сетевой службе:
1. в качестве формата сообщений для протокола взаимодействия пиров нужно использо-
вать JSON;
2. полная проверка данных, получаемых из сети;
3. устойчивость к обрыву соединения;
4. ведение подробного лога всех ошибок, а также других важных событий (установка и
завершение соединения с соседним пиром, приём и передача сообщений, и т.п.).

## Вариант 1. Чат (кольцо)
- Топология: кольцевой список.
- Информация, известная пиру при запуске: его IP-адрес и порт, а
также IP-адрес и порт следующего пира в кольцевом списке
(следующий пир не обязан быть заранее запущен).
- Описание службы: каждый пир имеет графический
пользовательский интерфейс (или web-интерфейс), через который
осуществляется просмотр и передача сообщений в чате.

## Пример работы

### Peer 1 (bobby) `/p1.sh`
```
Starting peer 1 (6001 -> 6002)
'm' - message, 'q' - quit
15:58:43.214573 INF ~ Server listening at 127.0.0.1:6001
15:58:44.217442 DBG ~ Clt could not connect to localhost:6002. retry 1 of 10
15:58:45.219042 DBG ~ Clt could not connect to localhost:6002. retry 2 of 10
15:58:46.223741 DBG ~ Clt could not connect to localhost:6002. retry 3 of 10
15:58:47.227559 DBG ~ Clt could not connect to localhost:6002. retry 4 of 10
15:58:48.232006 INF ~ Clt opening connection (127.0.0.1:59783 -> 127.0.0.1:6002)
command:
15:58:55.544829 DBG ~ Server handling connection (127.0.0.1:6001 <- 127.0.0.1:59793)
15:59:17.514110 INF ~ Incoming message
   donald: hello there
15:59:27.671380 INF ~ Incoming message
   donald: how r u
15:59:40.898398 INF ~ Incoming message
   eddy: fine thx
m
Type your message:
fck that im leaving
command:
q
15:59:56.692135 INF ~ Server is shutting down
```

### Peer 2 (donald) `./p2.sh`
```
Starting peer 2 (6002 -> 6003)
'm' - message, 'q' - quit
15:58:47.750252 INF ~ Server listening at 127.0.0.1:6002
15:58:48.232109 DBG ~ Server handling connection (127.0.0.1:6002 <- 127.0.0.1:59783)
15:58:48.754458 DBG ~ Clt could not connect to localhost:6003. retry 1 of 10
15:58:49.758572 DBG ~ Clt could not connect to localhost:6003. retry 2 of 10
15:58:50.760015 DBG ~ Clt could not connect to localhost:6003. retry 3 of 10
15:58:51.760979 DBG ~ Clt could not connect to localhost:6003. retry 4 of 10
15:58:52.764953 DBG ~ Clt could not connect to localhost:6003. retry 5 of 10
15:58:53.769181 DBG ~ Clt could not connect to localhost:6003. retry 6 of 10
15:58:54.773309 INF ~ Clt opening connection (127.0.0.1:59792 -> 127.0.0.1:6003)
command:
m
Type your message:
hello there
command:
m
Type your message:
how r u
command:
15:59:41.903150 INF ~ Incoming message
   eddy: fine thx
15:59:53.133607 INF ~ Incoming message
   bobby: fck that im leaving
15:59:56.692620 INF ~
   _: User bobby has left
15:59:56.692833 DBG ~ Connection to prev. peer probably lost
m
Type your message:
see ya
command:
q
16:00:21.805226 INF ~ Server is shutting down
```

### Peer 3 (eddy) `./p3.sh`
```
Starting peer 3 (6003 -> 6001)
'm' - message, 'q' - quit
15:58:54.540612 INF ~ Server listening at 127.0.0.1:6003
15:58:54.773425 DBG ~ Server handling connection (127.0.0.1:6003 <- 127.0.0.1:59792)
15:58:55.544748 INF ~ Clt opening connection (127.0.0.1:59793 -> 127.0.0.1:6001)
command:
15:59:16.508106 INF ~ Incoming message
   donald: hello there
15:59:26.667360 INF ~ Incoming message
   donald: how r u
m
Type your message:
fine thx
command:
15:59:54.138392 INF ~ Incoming message
   bobby: fck that im leaving
15:59:56.693346 INF ~
   _: User bobby has left
16:00:17.676548 INF ~ Incoming message
   donald: see ya
16:00:21.805643 INF ~
   _: User donald has left
16:00:21.805747 DBG ~ Connection to prev. peer probably lost
q
16:00:26.055126 INF ~ Server is shutting down
```

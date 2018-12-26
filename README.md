# Print Redis Topology

Print slaves of redis masters

## redis.txt

List host and port in files each line, separated host and port by space.

```
host1 port1
host2 port2
host3 port3
host4 port4
```

## Usage

```
cat redis.txt | redis-topology -a <auth>
```

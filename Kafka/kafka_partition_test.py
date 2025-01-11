from kafka import KafkaConsumer
from kafka import TopicPartition

# To consume latest messages and auto-commit offsets
#consumer = KafkaConsumer('my-topic3',
#                         group_id='my-group',
#                         bootstrap_servers=['192.168.150.80:9092'])

# To consume messages from a specific PARTITION  [ FIX ]
consumer = KafkaConsumer(bootstrap_servers='localhost:29092')
consumer.assign([TopicPartition('test', 2)])

for message in consumer:
    # message value and key are raw bytes -- decode if necessary!
    # e.g., for unicode: `message.value.decode('utf-8')`
    print ("Topic= %s : Partition= %d : Offset= %d: key= %s value= %s" % (message.topic, message.partition,
                                          message.offset, message.key,
                                          message.value))
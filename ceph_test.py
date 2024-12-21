import rados, sys

cluster = rados.Rados(conffile='ceph.conf')
print("\nlibrados version: {}".format(str(cluster.version())))
print("Will attempt to connect to: {}".format(str(cluster.conf_get('mon host'))))

cluster.connect()
print("\nCluster ID: {}".format(cluster.get_fsid()))

print("\n\nCluster Statistics")
print("==================")
cluster_stats = cluster.get_cluster_stats()

for key, value in cluster_stats.items():
     print(key, value)

print("\n\nPool Operations")
print("===============")

print("\nAvailable Pools")
print("----------------")
pools = cluster.list_pools()

for pool in pools:
     print(pool)

print("\nCreate 'test' Pool")
print("------------------")
cluster.create_pool('test')

print("\nPool named 'test' exists: {}".format(str(cluster.pool_exists('test'))))
print("\nVerify 'test' Pool Exists")
print("-------------------------")
pools = cluster.list_pools()

for pool in pools:
     print(pool)

ioctx = cluster.open_ioctx('test')
print("\nWriting object 'hw' with contents 'Hello World!' to pool 'test'.")
ioctx.write_full("hw", b"Hello World!")

print("\n\nContents of object 'hw'\n------------------------\n")
print(ioctx.read("hw"))

print("\nRemoving object 'hw'")
ioctx.remove_object("hw")

print("\n\nWriting XATTR 'lang' with value 'en_US' to object 'hw'")
ioctx.set_xattr("hw", "lang", b"en_US")

print("\n\nGetting XATTR 'lang' from object 'hw'\n")
print(ioctx.get_xattr("hw", "lang"))

object_iterator = ioctx.list_objects()

while True :

        try :
                rados_object = object_iterator.__next__()
                print("Object contents = {}".format(rados_object.read()))

        except StopIteration :
                break

print("\nClosing the connection.")
ioctx.close()

print("\nDelete 'test' Pool")
print("------------------")
cluster.delete_pool('test')
print("\nPool named 'test' exists: {}".format(str(cluster.pool_exists('test'))))

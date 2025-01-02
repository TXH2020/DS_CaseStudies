# Ceph

I had a ceph cluster running on my Debian VM (thanks to microceph). I had to locate ceph.conf and ceph.client.admin.keyring. I then had to add the following line under global section of ceph.conf.

``keyring = /var/snap/microceph/1228/conf/ceph.client.admin.keyring``

The python file for testing ceph client must be run as sudo.

While testing the python file, I found that pool delete didnt work.

I wanted to delete the pool from the CLI so I used the following commands:

``sudo ceph osd pool rm test``

Output of the above command orders to issue the command as follows:

``sudo ceph osd pool rm test test --yes-i-really-really-mean-it``

However, it said the pool deletion was disabled; to enable mon_allow_pool_delete had to be set to true. I added the line to ceph.conf and rebooted but it didnt work. Finally, I tried the following command:

``sudo ceph tell mon.* injectargs --mon_allow_pool_delete true``, following which the "test" pool was deleted. Even the Python code for pool deletion worked.

Concepts learnt: Ceph architecture, ceph client through librados python and ceph admin CLI, File vs Objects, CAP theorem, High Availability, Consistency Models(Linearizability, Serializability, Causal Consistency, Eventual Consistency), Ceph Papers Analysis(Ceph, CRUSH and RADOS), points related to bluestore and Ceph alternatives(Lustre and Gluster). 

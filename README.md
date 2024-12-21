# Ceph

I had a ceph cluster running on my Debian VM (thanks to microceph). I had to locate ceph.conf and ceph.client.admin.keyring. I then had to add the following line under global section of ceph.conf.

``keyring = /var/snap/microceph/1228/conf/ceph.client.admin.keyring``

The python file for testing ceph client must be run as sudo.

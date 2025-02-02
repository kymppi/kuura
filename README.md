# Kuura - Authentication with Great M2M support

> Kuura (finnish) translated means frost that paints the windows during wintertime

## Usage

Generate local keyfile for development: `openssl rand -out ./.kek 32`
Prime: `openssl dhparam -text -out dhparam.pem 3072`

### Example Prime

```
    DH Parameters: (3072 bit)
    P:
        00:b1:4c:be:b5:82:6b:34:e3:07:57:14:52:0d:e2:
        af:61:58:85:f2:44:35:8e:49:8a:04:de:5d:ea:9d:
        79:aa:14:2f:06:24:23:92:61:bb:23:09:fa:a2:50:
        a9:c4:b5:62:29:28:2f:6a:d7:ef:3b:44:c5:95:21:
        f3:2b:30:c6:2d:05:7c:25:b7:f7:99:26:18:a3:d1:
        32:93:90:ea:a0:c1:c1:2a:13:29:01:01:d7:7a:cd:
        8d:39:69:55:68:68:a8:b4:84:2a:28:cf:29:10:c4:
        31:ef:d3:da:63:d6:1e:5c:6f:03:2f:74:5f:53:99:
        96:15:7b:c6:b5:f6:bf:8d:7a:3f:72:87:95:0d:84:
        fe:c7:d5:22:7e:d4:6a:15:20:65:72:ac:d3:70:c8:
        ae:80:b9:a2:8b:69:38:d1:d8:f8:9f:64:02:c8:f6:
        4d:46:45:95:06:50:6e:6e:2c:51:b4:3d:df:34:41:
        48:24:3a:6b:82:c4:09:ef:9a:ec:54:0c:26:ef:f0:
        d3:12:4c:08:e4:9e:52:f2:d8:fd:32:ac:b1:e6:da:
        c5:c5:80:11:01:53:c4:46:31:d3:24:ac:ba:e6:52:
        28:3c:72:58:d9:99:cd:38:be:fb:99:68:90:6e:22:
        1d:7f:aa:36:67:09:97:2b:2a:45:c0:73:63:03:e8:
        4f:84:8f:fe:d4:35:f2:f4:18:5a:b7:0f:de:27:16:
        47:bf:26:ae:bf:86:f8:ac:72:11:b9:65:ea:95:92:
        98:cf:ae:ff:20:6a:60:c5:5f:35:34:ca:05:ea:f7:
        12:32:76:2e:c5:43:98:f1:cb:55:40:02:f9:01:d0:
        af:df:b3:ad:84:d4:a2:dc:e1:4b:6a:fb:0e:41:97:
        a9:a6:17:34:2a:d8:03:10:f5:46:07:62:e5:88:32:
        51:d6:64:ab:e2:d8:e9:26:78:b2:72:3e:9e:b7:a2:
        8a:e1:d5:5e:fe:29:87:61:1a:95:06:57:f2:63:98:
        d4:bf:5e:be:cf:a2:4b:ce:c5:97
    G:    2 (0x2)
    recommended-private-length: 275 bits
-----BEGIN DH PARAMETERS-----
MIIBjAKCAYEAsUy+tYJrNOMHVxRSDeKvYViF8kQ1jkmKBN5d6p15qhQvBiQjkmG7
Iwn6olCpxLViKSgvatfvO0TFlSHzKzDGLQV8Jbf3mSYYo9Eyk5DqoMHBKhMpAQHX
es2NOWlVaGiotIQqKM8pEMQx79PaY9YeXG8DL3RfU5mWFXvGtfa/jXo/coeVDYT+
x9UiftRqFSBlcqzTcMiugLmii2k40dj4n2QCyPZNRkWVBlBubixRtD3fNEFIJDpr
gsQJ75rsVAwm7/DTEkwI5J5S8tj9Mqyx5trFxYARAVPERjHTJKy65lIoPHJY2ZnN
OL77mWiQbiIdf6o2ZwmXKypFwHNjA+hPhI/+1DXy9Bhatw/eJxZHvyauv4b4rHIR
uWXqlZKYz67/IGpgxV81NMoF6vcSMnYuxUOY8ctVQAL5AdCv37OthNSi3OFLavsO
QZepphc0KtgDEPVGB2LliDJR1mSr4tjpJniycj6et6KK4dVe/imHYRqVBlfyY5jU
v16+z6JLzsWXAgECAgIBEw==
-----END DH PARAMETERS-----
```

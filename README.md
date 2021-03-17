# coredns-netbox-plugin

This plugin gets an A record from NetBox[1]. It uses the REST API of netxbox
to ask for a an IP address of a hostname:

https://netbox.example.org/api/dcim/devices/?name=example-vm-host


```json
{
  "count": 1,
  "next": null,
  "previous": null,
  "results": [
    {
      "id": 106,
      "name": "example",
      "display_name": "example",
      "device_type": {
        "id": 4,
        "url": "https://netbox.example.org/api/dcim/device-types/4/",
        "manufacturer": {
          "id": 9,
          "url": "https://netbox.example.org/api/dcim/manufacturers/9/",
          "name": "IBM",
          "slug": "ibm"
        },
        "model": "x3350 M4",
        "slug": "x3350-m4",
        "display_name": "IBM x3350 M4"
      },
      "device_role": {
        "id": 8,
        "url": "https://netbox.example.org/api/dcim/device-roles/8/",
        "name": "Server",
        "slug": "server"
      },
      "tenant": null,
      "platform": {
        "id": 5,
        "url": "https://netbox.example.org/api/dcim/platforms/5/",
        "name": "Linux",
        "slug": "linux"
      },
      "serial": "",
      "asset_tag": null,
      "site": {
        "id": 4,
        "url": "https://netbox.example.org/api/dcim/sites/4/",
        "name": "site1",
        "slug": "site1"
      },
      "rack": {
        "id": 22,
        "url": "https://netbox.example.org/api/dcim/racks/22/",
        "name": "site1 Rack 1",
        "display_name": "site1 Rack 1"
      },
      "position": 30,
      "face": {
        "value": 0,
        "label": "Front"
      },
      "parent_device": null,
      "status": {
        "value": 1,
        "label": "Active"
      },
      "primary_ip": {
        "id": 209,
        "url": "https://netbox.example.org/api/ipam/ip-addresses/209/",
        "family": 4,
        "address": "172.16.50.5/24"
      },
      "primary_ip4": {
        "id": 209,
        "url": "https://netbox.example.org/api/ipam/ip-addresses/209/",
        "family": 4,
        "address": "172.16.50.5/24"
      },
      "primary_ip6": null,
      "cluster": null,
      "virtual_chassis": null,
      "vc_position": null,
      "vc_priority": null,
      "comments": "",
      "local_context_data": null,
      "tags": [],
      "custom_fields": {},
      "created": "2021-03-05",
      "last_updated": "2021-03-15T11:57:55.886871Z"
    }
  ]
}
```

## Usage

To activate the plugin you need to compile CoreDNS with the plugin added
to `plugin.cfg`

```
netbox:github.com/cosandr/coredns-netbox-plugin
```

Then add it to Corefile:

```
. {
   netbox {
      token <YOU-NETBOX-API-TOKEN>
      url <https://netbox.example.org>
      localCacheDuration <The duration to keep each entry locally before querying netbox again. Use go `time.Duration` notation>
   }
}
```

The config parameters are mandatory.
## Developing locally

You can test the plugin functionallity with CoreDNS by adding the following to
`go.mod` in the source code directory of coredns.

```
replace github.com/cosandr/coredns-netbox-plugin => <path-to-you-local-copy>
```

Testing against a remote instance of netbox is possible with SSH port forwarding:

```
Host YourHost
   Hostname 10.0.0.91
   ProxyJump YourJumpHost
   LocalForward 18443 192.168.1.128:8443
```

## Credits

This plugin is heavily based on the code of the redis-plugin for CoreDNS.


[1]: https://netbox.readthedocs.io/en/stable/

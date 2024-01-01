
## Node

The management node is a technology solution used to logically group network devices for unified management and control. By grouping network devices, the administrator can more easily manage the network devices and also facilitate maintenance and management of the network devices.

## CPE 

CPE (Customer Premises Equipment) refers to network hardware located at the customer's location, such as a residence or business, used to provide a connection to the service provider's network. These devices include modems, routers, switches and other network equipment used to provide internet and other communication services to customers. In TeamsACS, CPE functions as the TR069 protocol client and communicates with TeamsACS.

Mikrotik's RouterOS router, produced by Mikrotik, is a good example and a commonly used CPE device that provides Internet and other communication services. By communicating with TeamsACS through the TR069 protocol, administrators can manage and control CPE devices more easily.

## TR069 Config 

In TeamsACS, TR069 Config is mainly Configuration data of Configuration Change (3 Vendor Configuration File) type, which is sent to CPE through 'Download RPC method'

Mikrotik devices are more flexible in dealing with Configuration Change (3 Vendor Configuration File) configuration data. The configuration it receives is the RouterOS script, through which more operations can be performed, such as configuring the router, setting up the firewall, etc. Mikrotik devices are a good choice to use because of their flexibility.

The following is a description quoting the official [Mikrotik Wiki](https://help.mikrotik.com/docs/display/ROS/TR-069)

`Configuration Change (3 Vendor Configuration File)`

> The same Download RPC can be used to perform complete configuration overwrite (as intended by standard) OR configuration alteration (when URL's filename extension is ".alter").

`Alter configuration`

> RouterOS has a lot of configuration attributes and not everything can be ported to CWMP Parameters, that's why RouterOS provides a possibility to execute its powerful scripting language to configure any attribute. A configuration alteration (which is really a regular script execution) can be performed using Download RPC FileType="3 Vendor Configuration File" with downloadable file extension ".alter". This powerful feature can be used to configure any ROS attributes which are not available through CWMP Parameters.

`Overwrite all configurations`

> Full ROS configuration overwrite can be performed using Download RPC FileType="3 Vendor Configuration File" with any URL file name (except with ".alter" extension).

### TR069 config session

The TR069 config can be triggered manually at any time in the TeamsACS system. A full execution log is recorded for each release so that the user can view the status of the release at any time. This increases the transparency and traceability of the system and allows managers to monitor and manage the issuance of TR069 configs more effectively.


## Firmware config

In TeamsACS, Firmware config is mainly configuration data of type (1 Firmware Upgrade Image), which is sent to CPE through "download RPC method"

The following is a description quoting the official [Mikrotik Wiki](https://help.mikrotik.com/docs/display/ROS/TR-069)


`RouterOS Update (1 Firmware Upgrade Image)`

> CWMP standard defines that CPE's firmware can be updated using Download RPC with FileType="1 Firmware Upgrade Image" and single URL of a downloadable file (HTTP and HTTPS are supported). Standard also states that downloaded file can be any type and vendor specific process can be applied to finish firmware update. Because MikroTik's update is package based (and also for extra flexibility), an XML file is used to describe firmware upgrade/downgrade. For now, XML configuration supports providing multiple URLs of files, which will be downloaded and applied similarly as regular RouterOS update through firmware/package file upload.

> An example of RouterOS bundle package and tr069-client package update (don't forget to also update tr069-client package). An XML file should be put on some HTTP server, which is accessible from CPE for download. Also, downloadable RouterOS package files should be accessible the same way (can be on any HTTP server). Using ACS execute Download RPC with URL pointing to XML file (e.g. "https://example.com/path/upgrade.xml") with contents:

```
<upgrade version="1" type="links">
   <config/>
   <links>
       <link>
          <url>https://example.com/routeros-mipsbe-X.Y.Z.npk</url>
       </link>
       <link>
          <url>https://example.com/tr069-client-X.Y.Z-mipsbe.npk</url>
       </link>
   </links>
</upgrade>
```

> CPE will download XML, parse/validate its contents, download files from provided URLs and try to upgrade. The result will be reported with TransferComplete RPC.


## Factoryreset config


The following is a description quoting the official [Mikrotik Wiki](https://help.mikrotik.com/docs/display/ROS/TR-069)


`RouterOS default configuration change (X MIKROTIK Factory Configuration File)`

> This vendor specific FileType allows the change of the RouterOS default configuration script that is executed when /system reset-configuration command is executed (or the other means when router configuration is beeing reset).

## TR069 preset

In TeamsACS, a Tr069 preset is a pre-configured TR069 RPC operation described in a yaml format file, which includes the RPC method to be executed, parameters, error handling methods, etc. Tr069 presets can be executed manually or triggered by TR069 specific events such as a boot event, or by a backend timed task in TeamsACS.

### TR069 preset task

When the TR069 preset execution is triggered, the TeamsACS system creates a task to track the execution of the TR069 preset. During execution, the system records the content and status of the execution, making it easy for managers to monitor and evaluate the execution process. This helps managers to have a better understanding of the status of the configuration being issued and to be able to address any issues in a timely manner.





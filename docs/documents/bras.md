## Introduction to BRAS

Broadband Remote Access Server (BRAS) is a high-capacity network device typically located in a service provider's (ISP's) network. It plays a crucial role in delivering high-speed internet access services, particularly in DSL or cable internet services.

## Role of BRAS

A BRAS is primarily used to manage user sessions and routing information, enabling users to access the internet. It serves as an intermediary between the user access network (like DSL or cable) and the service provider's core network. Upon connection of a user device (like a personal computer or a router) to the network, the BRAS carries out authentication, assigns IP addresses, and facilitates network access.

In the ToughRADIUS system, a VPE device assumes the role of a BRAS, delivering these critical functionalities.

### Functions of BRAS

Here are some of the primary functions of a BRAS:

User Authentication: The BRAS uses protocols like RADIUS or DIAMETER to verify the credentials of users and determine whether they should be allowed access to the network.

IP Address Management: Once a user is authenticated, the BRAS assigns one or more IP addresses to the user, enabling them to communicate within the network.

Session Management: The BRAS keeps track of each user's session, including the start and end times of the session, the amount of data transmitted, etc. This data can be used for billing, traffic management, and troubleshooting purposes.

Routing: The BRAS also handles the routing of user traffic, ensuring that data packets are accurately delivered from their source to their destination.

Quality of Service (QoS) Management: The BRAS can set and enforce QoS policies to ensure fair use of network resources and to meet the performance needs of different users and applications.

From the above, it is clear that the VPE, acting as a BRAS device in the ToughRADIUS system, is a key device for users to access the network and engage in network communication.


## VPE Model Definition

The VPE model is a Go language struct, defined as follows:

```golang
type NetVpe struct {
	ID         int64     `json:"id,string" form:"id"`            // Primary ID
	NodeId     int64     `json:"node_id,string" form:"node_id"`  // Node ID
	LdapId     int64     `json:"ldap_id,string" form:"ldap_id"`  // LDAP ID
	Name       string    `json:"name" form:"name"`               // Device name
	Identifier string    `json:"identifier" form:"identifier"`   // Device Identifier - RADIUS
	Hostname   string    `json:"hostname" form:"hostname"`       // Device host address
	Ipaddr     string    `json:"ipaddr" form:"ipaddr"`           // Device IP
	Secret     string    `json:"secret" form:"secret"`           // Device RADIUS Secret
	CoaPort    int       `json:"coa_port" form:"coa_port"`       // Device RADIUS COA port
	Model      string    `json:"model" form:"model"`             // Device model
	VendorCode string    `json:"vendor_code" form:"vendor_code"` // Device vendor code
	Status     string    `json:"status" form:"status"`           // Device status
	Tags       string    `json:"tags" form:"tags"`               // Tags
	Remark     string    `json:"remark" form:"remark"`           // Remark
	CreatedAt  time.Time `json:"created_at"`                     // Created at
	UpdatedAt  time.Time `json:"updated_at"`                     // Updated at
}

```


## Field explanations:

* ID: Unique identifier of the VPE device
* NodeId: Node ID
* LdapId: LDAP ID
* Name: Name of the VPE device
* Identifier: RADIUS identifier of the device
* Hostname: Host address of the device
* Ipaddr: IP address of the device
* Secret: RADIUS Secret of the device
* CoaPort: RADIUS COA port of the device
* Model: Model of the device
* VendorCode: Vendor code of the device
* Status: Status of the device
* Tags: Tags of the device
* Remark: Remarks about the device
* CreatedAt: When the device entry was created
* UpdatedAt: When the device entry was last updated


## VPE Creation

The creation of a VPE device primarily involves:

* Submit a form containing device information
* Validate necessary fields are not empty (e.g., name, VendorCode, and Identifier)
* Create the new VPE device in the database

<img width="1262" alt="image" src="https://github.com/talkincode/toughradius/assets/377938/e37804e4-e047-481e-a3b6-12d8f4d4c092">


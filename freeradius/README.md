# freeRADIUS integration


![freeradius-toughradius](https://github.com/talkincode/toughradius/assets/377938/f735d45d-3325-49e5-8b73-21c6205248e3)


TOUGHRADIUS integrates the FreeRADIUS API interface, extending its already comprehensive authentication capabilities and providing even more robust solutions to its clients. 
Our integration of the FreeRADIUS API allows for seamless integration with existing network infrastructures and enables us to offer a wider range of authentication options to meet the unique needs of our clients. 
Whether you require support for 802.1X, Wi-Fi, VPN, or other network access protocols, TOUGHRADIUS has you covered. With our advanced authentication capabilities and integration with FreeRADIUS, our clients can enjoy a secure, reliable, and efficient network management experience.


- FreeRadius enables the REST Module to interface with ToughRadius and uses HTTP parameters to pass user information
- Freeradius acts as the core of the RADIUS protocol processing
- Toughradius acts as a function for user management, billing, device management, device configuration management, and more
- Toughradius starts an HTTP server that listens for Freeradius requests and handles user authentication, billing, device management, and more

> Please refer to [Freeradius Official Documentation](https://networkradius.com/doc/3.0.10/raddb/mods-available/rest.html)

> [FreeRadius configuration case](https://github.com/talkincode/toughradius/tree/main/assets/freeradius)
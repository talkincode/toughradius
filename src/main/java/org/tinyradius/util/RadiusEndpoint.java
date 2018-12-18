/**
 * $Id: RadiusEndpoint.java,v 1.1 2005/09/07 22:19:01 wuttke Exp $
 * Created on 07.09.2005
 * @author glanz, Matthias Wuttke
 * @version $Revision: 1.1 $
 */
package org.tinyradius.util;

import java.net.InetSocketAddress;

/**
 * This class stores information about a Radius endpoint.
 * This includes the address of the remote endpoint and the shared secret
 * used for securing the communication.
 */
public class RadiusEndpoint {

	/**
	 * Constructs a RadiusEndpoint object.
	 * @param remoteAddress remote address (ip and port number)
	 * @param sharedSecret shared secret
	 */
	public RadiusEndpoint(InetSocketAddress remoteAddress, String sharedSecret) {
		this.endpointAddress = remoteAddress;
		this.sharedSecret = sharedSecret;
	}
	
	/**
	 * Returns the remote address.
	 * @return remote address
	 */
	public InetSocketAddress getEndpointAddress() {
		return endpointAddress;
	}
	
	/**
	 * Returns the shared secret.
	 * @return shared secret
	 */
	public String getSharedSecret() {
		return sharedSecret;
	}

	private InetSocketAddress endpointAddress;
    private String sharedSecret;

}

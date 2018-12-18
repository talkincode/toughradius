/**
 * $Id: RadiusProxyConnection.java,v 1.2 2005/10/11 14:18:27 wuttke Exp $
 * Created on 07.09.2005
 * @author glanz, Matthias Wuttke
 * @version $Revision: 1.2 $
 */
package org.tinyradius.proxy;

import org.tinyradius.packet.RadiusPacket;
import org.tinyradius.util.RadiusEndpoint;

/**
 * This class stores information about a proxied packet.
 * It contains two RadiusEndpoint objects representing the Radius client
 * and server, the port number the proxied packet arrived
 * at originally and the proxied packet itself.
 */
public class RadiusProxyConnection {

	/**
	 * Creates a RadiusProxyConnection object.
	 * @param radiusServer server endpoint
	 * @param radiusClient client endpoint
	 * @param port port the proxied packet arrived at originally 
	 */
	public RadiusProxyConnection(RadiusEndpoint radiusServer, RadiusEndpoint radiusClient, RadiusPacket packet, int port) {
		this.radiusServer = radiusServer;
		this.radiusClient = radiusClient;
		this.packet = packet;
		this.port = port;
	}
	
	/**
	 * Returns the Radius endpoint of the client.
	 * @return endpoint
	 */
	public RadiusEndpoint getRadiusClient() {
		return radiusClient;
	}
	
	/**
	 * Returns the Radius endpoint of the server.
	 * @return endpoint
	 */
	public RadiusEndpoint getRadiusServer() {
		return radiusServer;
	}
	
	/**
	 * Returns the proxied packet.
	 * @return packet 
	 */
	public RadiusPacket getPacket() {
		return packet;
	}
	
	/**
	 * Returns the port number the proxied packet arrived at
	 * originally. 
	 * @return port number
	 */
	public int getPort() {
		return port;
	}
	
	private RadiusEndpoint radiusServer;
	private RadiusEndpoint radiusClient;
	private int port;
	private RadiusPacket packet;
	
}

/**
 * $Id: RadiusClient.java,v 1.7 2005/11/10 10:20:21 wuttke Exp $
 * Created on 09.04.2005
 * 
 * @author Matthias Wuttke
 * @version $Revision: 1.7 $
 */
package org.tinyradius.util;

import java.io.ByteArrayInputStream;
import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.net.DatagramPacket;
import java.net.DatagramSocket;
import java.net.InetAddress;
import java.net.SocketException;
import java.net.SocketTimeoutException;
import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;
import org.tinyradius.packet.AccessRequest;
import org.tinyradius.packet.AccountingRequest;
import org.tinyradius.packet.RadiusPacket;

/**
 * This object represents a simple Radius client which communicates with
 * a specified Radius server. You can use a single instance of this object
 * to authenticate or account different users with the same Radius server
 * as long as you authenticate/account one user after the other. This object
 * is thread safe, but only opens a single socket so operations using this
 * socket are synchronized to avoid confusion with the mapping of request
 * and result packets.
 */
public class RadiusClient {

	/**
	 * Creates a new Radius client object for a special Radius server.
	 * 
	 * @param hostName
	 *            host name of the Radius server
	 * @param sharedSecret
	 *            shared secret used to secure the communication
	 */
	public RadiusClient(String hostName, String sharedSecret) {
		setHostName(hostName);
		setSharedSecret(sharedSecret);
	}

	/**
	 * Constructs a Radius client for the given Radius endpoint.
	 * 
	 * @param client
	 *            Radius endpoint
	 */
	public RadiusClient(RadiusEndpoint client) {
		this(client.getEndpointAddress().getAddress().getHostAddress(), client.getSharedSecret());
	}

	/**
	 * Authenticates a user via PAP.
	 * 
	 * @param userName
	 *            user name
	 * @param password
	 *            password
	 * @return true if authentication is successful, false otherwise
	 * @exception RadiusException
	 *                malformed packet
	 * @exception IOException
	 *                communication error (after getRetryCount()
	 *                retries)
	 */
	public synchronized boolean authenticate(String userName, String password) throws IOException, RadiusException {
		return authenticate(userName, password, authProtocol);
	}

	/**
	 * Authenticates a user.
	 * 
	 * @param userName
	 *            user name
	 * @param password
	 *            password
	 * @param protocol
	 *            either {@link AccessRequest#AUTH_PAP} or {@link AccessRequest#AUTH_CHAP}
	 * @return true if authentication is successful, false otherwise
	 * @exception RadiusException
	 *                malformed packet
	 * @exception IOException
	 *                communication error (after getRetryCount()
	 *                retries)
	 */
	public synchronized boolean authenticate(String userName, String password, String protocol) throws IOException, RadiusException {
		AccessRequest request = new AccessRequest(userName, password);
		request.setAuthProtocol(protocol);
		RadiusPacket response = authenticate(request);
		return response.getPacketType() == RadiusPacket.ACCESS_ACCEPT;
	}

	/**
	 * Sends an Access-Request packet and receives a response
	 * packet.
	 * 
	 * @param request
	 *            request packet
	 * @return Radius response packet
	 * @exception RadiusException
	 *                malformed packet
	 * @exception IOException
	 *                communication error (after getRetryCount()
	 *                retries)
	 */
	public synchronized RadiusPacket authenticate(AccessRequest request) throws IOException, RadiusException {
		if (logger.isInfoEnabled())
			logger.info("send Access-Request packet: " + request);

		RadiusPacket response = communicate(request, getAuthPort());
		if (logger.isInfoEnabled())
			logger.info("received packet: " + response);

		return response;
	}

	/**
	 * Sends an Accounting-Request packet and receives a response
	 * packet.
	 * 
	 * @param request
	 *            request packet
	 * @return Radius response packet
	 * @exception RadiusException
	 *                malformed packet
	 * @exception IOException
	 *                communication error (after getRetryCount()
	 *                retries)
	 */
	public synchronized RadiusPacket account(AccountingRequest request) throws IOException, RadiusException {
		if (logger.isInfoEnabled())
			logger.info("send Accounting-Request packet: " + request);

		RadiusPacket response = communicate(request, getAcctPort());
		if (logger.isInfoEnabled())
			logger.info("received packet: " + response);

		return response;
	}

	/**
	 * Closes the socket of this client.
	 */
	public void close() {
		if (serverSocket != null)
			serverSocket.close();
	}

	/**
	 * Returns the Radius server auth port.
	 * 
	 * @return auth port
	 */
	public int getAuthPort() {
		return authPort;
	}

	/**
	 * Sets the auth port of the Radius server.
	 * 
	 * @param authPort
	 *            auth port, 1-65535
	 */
	public void setAuthPort(int authPort) {
		if (authPort < 1 || authPort > 65535)
			throw new IllegalArgumentException("bad port number");
		this.authPort = authPort;
	}

	/**
	 * Returns the host name of the Radius server.
	 * 
	 * @return host name
	 */
	public String getHostName() {
		return hostName;
	}

	/**
	 * Sets the host name of the Radius server.
	 * 
	 * @param hostName
	 *            host name
	 */
	public void setHostName(String hostName) {
		if (hostName == null || hostName.length() == 0)
			throw new IllegalArgumentException("host name must not be empty");
		this.hostName = hostName;
	}

	/**
	 * Returns the retry count for failed transmissions.
	 * 
	 * @return retry count
	 */
	public int getRetryCount() {
		return retryCount;
	}

	/**
	 * Sets the retry count for failed transmissions.
	 * 
	 * @param retryCount
	 *            retry count, >0
	 */
	public void setRetryCount(int retryCount) {
		if (retryCount < 1)
			throw new IllegalArgumentException("retry count must be positive");
		this.retryCount = retryCount;
	}

	/**
	 * Returns the secret shared between server and client.
	 * 
	 * @return shared secret
	 */
	public String getSharedSecret() {
		return sharedSecret;
	}

	/**
	 * Sets the secret shared between server and client.
	 * 
	 * @param sharedSecret
	 *            shared secret
	 */
	public void setSharedSecret(String sharedSecret) {
		if (sharedSecret == null || sharedSecret.length() == 0)
			throw new IllegalArgumentException("shared secret must not be empty");
		this.sharedSecret = sharedSecret;
	}

	/**
	 * Returns the socket timeout.
	 * 
	 * @return socket timeout, ms
	 */
	public int getSocketTimeout() {
		return socketTimeout;
	}

	/**
	 * Sets the socket timeout
	 * 
	 * @param socketTimeout
	 *            timeout, ms, >0
	 * @throws SocketException
	 */
	public void setSocketTimeout(int socketTimeout) throws SocketException {
		if (socketTimeout < 1)
			throw new IllegalArgumentException("socket tiemout must be positive");
		this.socketTimeout = socketTimeout;
		if (serverSocket != null)
			serverSocket.setSoTimeout(socketTimeout);
	}

	/**
	 * Sets the Radius server accounting port.
	 * 
	 * @param acctPort
	 *            acct port, 1-65535
	 */
	public void setAcctPort(int acctPort) {
		if (acctPort < 1 || acctPort > 65535)
			throw new IllegalArgumentException("bad port number");
		this.acctPort = acctPort;
	}

	/**
	 * Returns the Radius server accounting port.
	 * 
	 * @return acct port
	 */
	public int getAcctPort() {
		return acctPort;
	}

	/**
	 * Set the Radius authentication protocol.
	 * 
	 * @see AccessRequest#AUTH_CHAP
	 * @see AccessRequest#AUTH_PAP
	 * 
	 * @param protocol
	 *            the protocol, PAP or CHAP
	 */
	public void setAuthProtocol(String protocol) {
		this.authProtocol = protocol;
	}

	/**
	 * Sends a Radius packet to the server and awaits an answer.
	 * 
	 * @param request
	 *            packet to be sent
	 * @param port
	 *            server port number
	 * @return response Radius packet
	 * @exception RadiusException
	 *                malformed packet
	 * @exception IOException
	 *                communication error (after getRetryCount()
	 *                retries)
	 */
	public RadiusPacket communicate(RadiusPacket request, int port) throws IOException, RadiusException {
		DatagramPacket packetIn = new DatagramPacket(new byte[RadiusPacket.MAX_PACKET_LENGTH], RadiusPacket.MAX_PACKET_LENGTH);
		DatagramPacket packetOut = makeDatagramPacket(request, port);

		DatagramSocket socket = getSocket();
		for (int i = 1; i <= getRetryCount(); i++) {
			try {
				socket.send(packetOut);
				socket.receive(packetIn);
				return makeRadiusPacket(packetIn, request);
			} catch (IOException ioex) {
				if (i == getRetryCount()) {
					if (logger.isErrorEnabled()) {
						if (ioex instanceof SocketTimeoutException)
							logger.error("communication failure (timeout), no more retries");
						else
							logger.error("communication failure, no more retries", ioex);
					}
					throw ioex;
				}
				if (logger.isInfoEnabled())
					logger.info("communication failure, retry " + i);
				// TODO increase Acct-Delay-Time by getSocketTimeout()/1000
				// this changes the packet authenticator and requires packetOut to be
				// calculated again (call makeDatagramPacket)
			}
		}

		return null;
	}

	/**
	 * Sends the specified packet to the specified Radius server endpoint.
	 * 
	 * @param remoteServer
	 *            Radius endpoint consisting of server address,
	 *            port number and shared secret
	 * @param request
	 *            Radius packet to be sent
	 * @return received response packet
	 * @throws RadiusException
	 *             malformed packet
	 * @throws IOException
	 *             error while communication
	 */
	public static RadiusPacket communicate(RadiusEndpoint remoteServer, RadiusPacket request) throws RadiusException, IOException {
		RadiusClient rc = new RadiusClient(remoteServer);
		return rc.communicate(request, remoteServer.getEndpointAddress().getPort());
	}

	/**
	 * Returns the socket used for the server communication. It is
	 * bound to an arbitrary free local port number.
	 * 
	 * @return local socket
	 * @throws SocketException
	 */
	protected DatagramSocket getSocket() throws SocketException {
		if (serverSocket == null) {
			serverSocket = new DatagramSocket();
			serverSocket.setSoTimeout(getSocketTimeout());
		}
		return serverSocket;
	}

	/**
	 * Creates a datagram packet from a RadiusPacket to be send.
	 * 
	 * @param packet
	 *            RadiusPacket
	 * @param port
	 *            destination port number
	 * @return new datagram packet
	 * @throws IOException
	 */
	protected DatagramPacket makeDatagramPacket(RadiusPacket packet, int port) throws IOException {
		ByteArrayOutputStream bos = new ByteArrayOutputStream();
		packet.encodeRequestPacket(bos, getSharedSecret());
		byte[] data = bos.toByteArray();

		InetAddress address = InetAddress.getByName(getHostName());
		DatagramPacket datagram = new DatagramPacket(data, data.length, address, port);
		return datagram;
	}

	/**
	 * Creates a RadiusPacket from a received datagram packet.
	 * 
	 * @param packet
	 *            received datagram
	 * @param request
	 *            Radius request packet
	 * @return RadiusPacket object
	 */
	protected RadiusPacket makeRadiusPacket(DatagramPacket packet, RadiusPacket request) throws IOException, RadiusException {
		ByteArrayInputStream in = new ByteArrayInputStream(packet.getData());
		return RadiusPacket.decodeResponsePacket(in, getSharedSecret(), request);
	}

	private int authPort = 1812;
	private int acctPort = 1813;
	private String hostName = null;
	private String sharedSecret = null;
	private DatagramSocket serverSocket = null;
	private int retryCount = 3;
	private int socketTimeout = 3000;
	private String authProtocol = AccessRequest.AUTH_PAP;
	private static Log logger = LogFactory.getLog(RadiusClient.class);

}

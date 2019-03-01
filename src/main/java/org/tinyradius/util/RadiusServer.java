/**
 * $Id: RadiusServer.java,v 1.11 2008/04/24 05:22:50 wuttke Exp $
 * Created on 09.04.2005
 *
 * @author Matthias Wuttke
 * @version $Revision: 1.11 $
 */
package org.tinyradius.util;

import java.io.ByteArrayInputStream;
import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.net.DatagramPacket;
import java.net.DatagramSocket;
import java.net.InetAddress;
import java.net.InetSocketAddress;
import java.net.SocketException;
import java.net.SocketTimeoutException;
import java.util.Arrays;
import java.util.Iterator;
import java.util.LinkedList;
import java.util.List;
import java.util.concurrent.ExecutorService;

import org.tinyradius.attribute.RadiusAttribute;
import org.tinyradius.packet.AccessRequest;
import org.tinyradius.packet.AccountingRequest;
import org.tinyradius.packet.RadiusPacket;
import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;

/**
 * Implements a simple Radius server. This class must be subclassed to
 * provide an implementation for getSharedSecret() and getUserPassword().
 * If the server supports accounting, it must override
 * accountingRequestReceived().
 */
public abstract class RadiusServer {

	/**
	 * Define this executor in child class to make packet processing be made in separate threads
	 */
	protected ExecutorService executor = null;

	/**
	 * Returns the shared secret used to communicate with the client with the
	 * passed IP address or null if the client is not allowed at this server.
	 *
	 * @param client
	 *            IP address and port number of client
	 * @return shared secret or null
	 */
	public abstract String getSharedSecret(InetSocketAddress client);

	/**
	 * Returns the shared secret used to communicate with the client with the
	 * passed IP address and the received packet data or null if the client 
	 * is not allowed at this server.
	 *
	 * for compatiblity this standard implementation just call the getSharedSecret(InetSocketAddress) method
	 * and should be overrived when necessary
	 *
	 * @param client
	 *            IP address and port number of client
	 * @param packet
	 *            packet received from client, the packettype comes as RESERVED, 
	 *	      because for some packets the secret is necessary for decoding
	 * @return shared secret or null
	 */
	public String getSharedSecret(InetSocketAddress client, RadiusPacket packet) {
		return getSharedSecret(client);
	}

	/**
	 * Returns the password of the passed user. Either this
	 * method or accessRequestReceived() should be overriden.
	 *
	 * @param userName
	 *            user name
	 * @return plain-text password or null if user unknown
	 */
	public abstract String getUserPassword(String userName);

	/**
	 * Constructs an answer for an Access-Request packet. Either this
	 * method or isUserAuthenticated should be overriden.
	 *
	 * @param accessRequest
	 *            Radius request packet
	 * @param client
	 *            address of Radius client
	 * @return response packet or null if no packet shall be sent
	 * @exception RadiusException
	 *                malformed request packet; if this
	 *                exception is thrown, no answer will be sent
	 */
	public RadiusPacket accessRequestReceived(AccessRequest accessRequest, InetSocketAddress client) throws RadiusException {
		String plaintext = getUserPassword(accessRequest.getUserName());
		int type = RadiusPacket.ACCESS_REJECT;
		if (plaintext != null && accessRequest.verifyPassword(plaintext))
			type = RadiusPacket.ACCESS_ACCEPT;

		RadiusPacket answer = new RadiusPacket(type, accessRequest.getPacketIdentifier());
		copyProxyState(accessRequest, answer);
		return answer;
	}

	/**
	 * Constructs an answer for an Accounting-Request packet. This method
	 * should be overriden if accounting is supported.
	 *
	 * @param accountingRequest
	 *            Radius request packet
	 * @param client
	 *            address of Radius client
	 * @return response packet or null if no packet shall be sent
	 * @exception RadiusException
	 *                malformed request packet; if this
	 *                exception is thrown, no answer will be sent
	 */
	public RadiusPacket accountingRequestReceived(AccountingRequest accountingRequest, InetSocketAddress client) throws RadiusException {
		RadiusPacket answer = new RadiusPacket(RadiusPacket.ACCOUNTING_RESPONSE, accountingRequest.getPacketIdentifier());
		copyProxyState(accountingRequest, answer);
		return answer;
	}

	/**
	 * Starts the Radius server.
	 *
	 * @param listenAuth
	 *            open auth port?
	 * @param listenAcct
	 *            open acct port?
	 */
	public void start(boolean listenAuth, boolean listenAcct) {
		if (listenAuth) {
			new Thread() {
				public void run() {
					setName("Radius Auth Listener");
					try {
						logger.info("starting RadiusAuthListener on port " + getAuthPort());
						listenAuth();
						logger.info("RadiusAuthListener is being terminated");
					}
					catch (Exception e) {
						e.printStackTrace();
						logger.fatal("auth thread stopped by exception", e);
					}
					finally {
						authSocket.close();
						logger.debug("auth socket closed");
					}
				}
			}.start();
		}

		if (listenAcct) {
			new Thread() {
				public void run() {
					setName("Radius Acct Listener");
					try {
						logger.info("starting RadiusAcctListener on port " + getAcctPort());
						listenAcct();
						logger.info("RadiusAcctListener is being terminated");
					}
					catch (Exception e) {
						e.printStackTrace();
						logger.fatal("acct thread stopped by exception", e);
					}
					finally {
						acctSocket.close();
						logger.debug("acct socket closed");
					}
				}
			}.start();
		}
	}

	/**
	 * Stops the server and closes the sockets.
	 */
	public void stop() {
		logger.info("stopping Radius server");
		closing = true;
		if (executor != null)
			executor.shutdown();
		if (authSocket != null)
			authSocket.close();
		if (acctSocket != null)
			acctSocket.close();
	}

	/**
	 * Returns the auth port the server will listen on.
	 *
	 * @return auth port
	 */
	public int getAuthPort() {
		return authPort;
	}

	/**
	 * Sets the auth port the server will listen on.
	 *
	 * @param authPort
	 *            auth port, 1-65535
	 */
	public void setAuthPort(int authPort) {
		if (authPort < 1 || authPort > 65535)
			throw new IllegalArgumentException("bad port number");
		this.authPort = authPort;
		this.authSocket = null;
	}

	/**
	 * Returns the socket timeout (ms).
	 *
	 * @return socket timeout
	 */
	public int getSocketTimeout() {
		return socketTimeout;
	}

	/**
	 * Sets the socket timeout.
	 *
	 * @param socketTimeout
	 *            socket timeout, >0 ms
	 * @throws SocketException
	 */
	public void setSocketTimeout(int socketTimeout) throws SocketException {
		if (socketTimeout < 1)
			throw new IllegalArgumentException("socket tiemout must be positive");
		this.socketTimeout = socketTimeout;
		if (authSocket != null)
			authSocket.setSoTimeout(socketTimeout);
		if (acctSocket != null)
			acctSocket.setSoTimeout(socketTimeout);
	}

	/**
	 * Sets the acct port the server will listen on.
	 *
	 * @param acctPort
	 *            acct port 1-65535
	 */
	public void setAcctPort(int acctPort) {
		if (acctPort < 1 || acctPort > 65535)
			throw new IllegalArgumentException("bad port number");
		this.acctPort = acctPort;
		this.acctSocket = null;
	}

	/**
	 * Returns the acct port the server will listen on.
	 *
	 * @return acct port
	 */
	public int getAcctPort() {
		return acctPort;
	}

	/**
	 * Returns the duplicate interval in ms.
	 * A packet is discarded as a duplicate if in the duplicate interval
	 * there was another packet with the same identifier originating from the
	 * same address.
	 *
	 * @return duplicate interval (ms)
	 */
	public long getDuplicateInterval() {
		return duplicateInterval;
	}

	/**
	 * Sets the duplicate interval in ms.
	 * A packet is discarded as a duplicate if in the duplicate interval
	 * there was another packet with the same identifier originating from the
	 * same address.
	 *
	 * @param duplicateInterval
	 *            duplicate interval (ms), >0
	 */
	public void setDuplicateInterval(long duplicateInterval) {
		if (duplicateInterval <= 0)
			throw new IllegalArgumentException("duplicate interval must be positive");
		this.duplicateInterval = duplicateInterval;
	}

	/**
	 * Returns the IP address the server listens on.
	 * Returns null if listening on the wildcard address.
	 *
	 * @return listen address or null
	 */
	public InetAddress getListenAddress() {
		return listenAddress;
	}

	/**
	 * Sets the address the server listens on.
	 * Must be called before start().
	 * Defaults to null, meaning listen on every
	 * local address (wildcard address).
	 *
	 * @param listenAddress
	 *            listen address or null
	 */
	public void setListenAddress(InetAddress listenAddress) {
		this.listenAddress = listenAddress;
	}

	/**
	 * Copies all Proxy-State attributes from the request
	 * packet to the response packet.
	 *
	 * @param request
	 *            request packet
	 * @param answer
	 *            response packet
	 */
	protected void copyProxyState(RadiusPacket request, RadiusPacket answer) {
		List proxyStateAttrs = request.getAttributes(33);
		for (Iterator i = proxyStateAttrs.iterator(); i.hasNext();) {
			RadiusAttribute proxyStateAttr = (RadiusAttribute) i.next();
			answer.addAttribute(proxyStateAttr);
		}
	}

	/**
	 * Listens on the auth port (blocks the current thread).
	 * Returns when stop() is called.
	 *
	 * @throws SocketException
	 * @throws InterruptedException
	 *
	 */
	protected void listenAuth() throws SocketException {
		listen(getAuthSocket());
	}

	/**
	 * Listens on the acct port (blocks the current thread).
	 * Returns when stop() is called.
	 *
	 * @throws SocketException
	 * @throws InterruptedException
	 */
	protected void listenAcct() throws SocketException {
		listen(getAcctSocket());
	}

	/**
	 * Listens on the passed socket, blocks until stop() is called.
	 *
	 * @param s
	 *            socket to listen on
	 */
	protected void listen(final DatagramSocket s) {
		while (true) {
			try {
				final DatagramPacket packetIn = new DatagramPacket(new byte[RadiusPacket.MAX_PACKET_LENGTH], RadiusPacket.MAX_PACKET_LENGTH);
				// receive packet
				try {
					logger.trace("about to call socket.receive()");
					s.receive(packetIn);
					if (logger.isDebugEnabled())
						logger.debug("receive buffer size = " + s.getReceiveBufferSize());
				}
				catch (SocketException se) {
					if (closing) {
						// end thread
						logger.info("got closing signal - end listen thread");
						return;
					}
					// retry s.receive()
					logger.error("SocketException during s.receive() -> retry", se);
					continue;
				}

				if (executor == null) {
					processRequest(s, packetIn);
				}
				else {
					executor.submit(new Runnable() {

						@Override
						public void run() {
							processRequest(s, packetIn);
						}

					});
				}
			}
			catch (SocketTimeoutException ste) {
				// this is expected behaviour
				logger.trace("normal socket timeout");
			}
			catch (IOException ioe) {
				// error while reading/writing socket
				logger.error("communication error", ioe);
			}
		}
	}


	/**
	 * Process a single received request
	 *
	 * @param s
	 *            socket to send response on
	 * @param packetIn
	 *		data packet 
	 */
	protected void processRequest(final DatagramSocket s, final DatagramPacket packetIn) {
		try {
			// check client
			final InetSocketAddress localAddress = (InetSocketAddress) s.getLocalSocketAddress();
			final InetSocketAddress remoteAddress = new InetSocketAddress(packetIn.getAddress(), packetIn.getPort());
			final String secret = getSharedSecret(remoteAddress, makeRadiusPacket(packetIn, "1234567890", RadiusPacket.RESERVED));
			if (secret == null) {
				if (logger.isInfoEnabled())
					logger.info("ignoring packet from unknown client " + remoteAddress + " received on local address " + localAddress);
				return;
			}

			// parse packet
			final RadiusPacket request = makeRadiusPacket(packetIn, secret, RadiusPacket.UNDEFINED);
			if (logger.isInfoEnabled())
				logger.info("received packet from " + remoteAddress + " on local address " + localAddress + ": " + request);

			// handle packet
			logger.trace("about to call RadiusServer.handlePacket()");
			final RadiusPacket response = handlePacket(localAddress, remoteAddress, request, secret);

			// send response
			if (response != null) {
				if (logger.isInfoEnabled())
					logger.info("send response: " + response);
				final DatagramPacket packetOut = makeDatagramPacket(response, secret, remoteAddress.getAddress(), packetIn.getPort(), request);
				s.send(packetOut);
			}
			else
				logger.info("no response sent");
		}
		catch (IOException ioe) {
			// error while reading/writing socket
			logger.error("communication error", ioe);
		}
		catch (RadiusException re) {
			// malformed packet
			logger.error("malformed Radius packet", re);
		}
	}

	/**
	 * Handles the received Radius packet and constructs a response.
	 *
	 * @param localAddress
	 *            local address the packet was received on
	 * @param remoteAddress
	 *            remote address the packet was sent by
	 * @param request
	 *            the packet
	 * @param sharedSecret
	 * @return response packet or null for no response
	 * @throws RadiusException
	 * @throws IOException
	 */
	protected RadiusPacket handlePacket(InetSocketAddress localAddress, InetSocketAddress remoteAddress, RadiusPacket request, String sharedSecret)
			throws RadiusException, IOException {
		RadiusPacket response = null;

		// check for duplicates
		if (!isPacketDuplicate(request, remoteAddress)) {
			if (localAddress.getPort() == getAuthPort()) {
				// handle packets on auth port
				if (request instanceof AccessRequest)
					response = accessRequestReceived((AccessRequest) request, remoteAddress);
				else
					logger.error("unknown Radius packet type: " + request.getPacketType());
			}
			else if (localAddress.getPort() == getAcctPort()) {
				// handle packets on acct port
				if (request instanceof AccountingRequest)
					response = accountingRequestReceived((AccountingRequest) request, remoteAddress);
				else
					logger.error("unknown Radius packet type: " + request.getPacketType());
			}
			else {
				// ignore packet on unknown port
			}
		}
		else
			logger.info("ignore duplicate packet");

		return response;
	}

	/**
	 * Returns a socket bound to the auth port.
	 *
	 * @return socket
	 * @throws SocketException
	 */
	protected DatagramSocket getAuthSocket() throws SocketException {
		if (authSocket == null) {
			if (getListenAddress() == null)
				authSocket = new DatagramSocket(getAuthPort());
			else
				authSocket = new DatagramSocket(getAuthPort(), getListenAddress());
			authSocket.setSoTimeout(getSocketTimeout());
		}
		return authSocket;
	}

	/**
	 * Returns a socket bound to the acct port.
	 *
	 * @return socket
	 * @throws SocketException
	 */
	protected DatagramSocket getAcctSocket() throws SocketException {
		if (acctSocket == null) {
			if (getListenAddress() == null)
				acctSocket = new DatagramSocket(getAcctPort());
			else
				acctSocket = new DatagramSocket(getAcctPort(), getListenAddress());
			acctSocket.setSoTimeout(getSocketTimeout());
		}
		return acctSocket;
	}

	/**
	 * Creates a Radius response datagram packet from a RadiusPacket to be send.
	 *
	 * @param packet
	 *            RadiusPacket
	 * @param secret
	 *            shared secret to encode packet
	 * @param address
	 *            where to send the packet
	 * @param port
	 *            destination port
	 * @param request
	 *            request packet
	 * @return new datagram packet
	 * @throws IOException
	 */
	protected DatagramPacket makeDatagramPacket(RadiusPacket packet, String secret, InetAddress address, int port, RadiusPacket request)
			throws IOException {
		ByteArrayOutputStream bos = new ByteArrayOutputStream();
		packet.encodeResponsePacket(bos, secret, request);
		byte[] data = bos.toByteArray();

		DatagramPacket datagram = new DatagramPacket(data, data.length, address, port);
		return datagram;
	}

	/**
	 * Creates a RadiusPacket for a Radius request from a received
	 * datagram packet.
	 *
	 * @param packet
	 *            received datagram
	 * @return RadiusPacket object
	 * @exception RadiusException
	 *                malformed packet
	 * @exception IOException
	 *                communication error (after getRetryCount()
	 *                retries)
	 */
	protected RadiusPacket makeRadiusPacket(DatagramPacket packet, String sharedSecret, int forceType) throws IOException, RadiusException {
		ByteArrayInputStream in = new ByteArrayInputStream(packet.getData());
		return RadiusPacket.decodeRequestPacket(in, sharedSecret, forceType);
	}

	/**
	 * Checks whether the passed packet is a duplicate.
	 * A packet is duplicate if another packet with the same identifier
	 * has been sent from the same host in the last time.
	 *
	 * @param packet
	 *            packet in question
	 * @param address
	 *            client address
	 * @return true if it is duplicate
	 */
	protected boolean isPacketDuplicate(RadiusPacket packet, InetSocketAddress address) {
		long now = System.currentTimeMillis();
		long intervalStart = now - getDuplicateInterval();

		byte[] authenticator = packet.getAuthenticator();

		synchronized (receivedPackets) {
			for (Iterator i = receivedPackets.iterator(); i.hasNext();) {
				ReceivedPacket p = (ReceivedPacket) i.next();
				if (p.receiveTime < intervalStart) {
					// packet is older than duplicate interval
					i.remove();
				}
				else {
					if (p.address.equals(address) && p.packetIdentifier == packet.getPacketIdentifier()) {
						if (authenticator != null && p.authenticator != null) {
							// packet is duplicate if stored authenticator is equal
							// to the packet authenticator
							return Arrays.equals(p.authenticator, authenticator);
						}
						// should not happen, packet is duplicate
						return true;
					}
				}
			}

			// add packet to receive list
			ReceivedPacket rp = new ReceivedPacket();
			rp.address = address;
			rp.packetIdentifier = packet.getPacketIdentifier();
			rp.receiveTime = now;
			rp.authenticator = authenticator;
			receivedPackets.add(rp);
		}

		return false;
	}

	private InetAddress listenAddress = null;
	private int authPort = 1812;
	private int acctPort = 1813;
	private DatagramSocket authSocket = null;
	private DatagramSocket acctSocket = null;
	private int socketTimeout = 3000;
	private List receivedPackets = new LinkedList();
	private long duplicateInterval = 30000; // 30 s
	protected transient boolean closing = false;
	private static Log logger = LogFactory.getLog(RadiusServer.class);

}

/**
 * This internal class represents a packet that has been received by
 * the server.
 */
class ReceivedPacket {

	/**
	 * The identifier of the packet.
	 */
	public int packetIdentifier;

	/**
	 * The time the packet was received.
	 */
	public long receiveTime;

	/**
	 * The address of the host who sent the packet.
	 */
	public InetSocketAddress address;

	/**
	 * Authenticator of the received packet.
	 */
	public byte[] authenticator;

}
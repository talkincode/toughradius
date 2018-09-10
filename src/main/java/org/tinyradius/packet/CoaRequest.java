package org.tinyradius.packet;

import java.security.MessageDigest;

import org.tinyradius.util.RadiusUtil;

/**
 * CoA packet. Thanks to Michael Krastev.
 * @author Michael Krastev <mkrastev@gmail.com>
 */
public class CoaRequest extends RadiusPacket {

	public CoaRequest() {
		this(COA_REQUEST);
	}
	public CoaRequest(final int type) {
		super(type, getNextPacketIdentifier()); 
	}
	
	/**
	 * @see AccountingRequest#updateRequestAuthenticator(String, int, byte[])
	 */
	protected byte[] updateRequestAuthenticator(String sharedSecret,
			int packetLength, byte[] attributes) {
		byte[] authenticator = new byte[16];
		for (int i = 0; i < 16; i++)
			authenticator[i] = 0;
		MessageDigest md5 = getMd5Digest();
		md5.reset();
		md5.update((byte) getPacketType());
		md5.update((byte) getPacketIdentifier());
		md5.update((byte) (packetLength >> 8));
		md5.update((byte) (packetLength & 0xff));
		md5.update(authenticator, 0, authenticator.length);
		md5.update(attributes, 0, attributes.length);
		md5.update(RadiusUtil.getUtf8Bytes(sharedSecret));
		return md5.digest();
	}
	
}

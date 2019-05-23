package org.tinyradius.packet;

import org.tinyradius.attribute.IntegerAttribute;
import org.tinyradius.attribute.RadiusAttribute;
import org.tinyradius.attribute.StringAttribute;
import org.tinyradius.util.RadiusException;
import org.tinyradius.util.RadiusUtil;

import java.security.MessageDigest;
import java.util.Iterator;
import java.util.List;

/**
 * This class represents a Radius packet of the type
 * "Accounting-Request".
 */
public class AccountingRequest extends RadiusPacket {


	/**
	 * Acct-Status-Type: Start
	 */
	public static final int ACCT_STATUS_TYPE_START = 1;

	/**
	 * Acct-Status-Type: Stop
	 */
	public static final int ACCT_STATUS_TYPE_STOP = 2;

	/**
	 * Acct-Status-Type: Interim Update/Alive
	 */
	public static final int ACCT_STATUS_TYPE_INTERIM_UPDATE = 3;

	/**
	 * Acct-Status-Type: Accounting-On
	 */
	public static final int ACCT_STATUS_TYPE_ACCOUNTING_ON = 7;

	/**
	 * Acct-Status-Type: Accounting-Off
	 */
	public static final int ACCT_STATUS_TYPE_ACCOUNTING_OFF = 8;

	/**
	 * Constructs an Accounting-Request packet to be sent to a Radius server.
	 * 
	 * @param userName
	 *            user name
	 * @param acctStatusType
	 *            ACCT_STATUS_TYPE_*
	 */
	public AccountingRequest(String userName, int acctStatusType) {
		super(ACCOUNTING_REQUEST, getNextPacketIdentifier());
		setUserName(userName);
		setAcctStatusType(acctStatusType);
	}

	/**
	 * Constructs an empty Accounting-Request to be received by a
	 * Radius client.
	 */
	public AccountingRequest() {
		super(ACCOUNTING_REQUEST);
	}

	public String getStatusTypeName(){
		switch (getAcctStatusType()){
			case ACCT_STATUS_TYPE_START:return "Start";
			case ACCT_STATUS_TYPE_INTERIM_UPDATE:return "Update";
			case ACCT_STATUS_TYPE_STOP:return "Stop";
			case ACCT_STATUS_TYPE_ACCOUNTING_ON:return "AcctOn";
			case ACCT_STATUS_TYPE_ACCOUNTING_OFF:return "AcctOff";
			default:return "Unknow";
		}
	}

	/**
	 * Sets the RadUser-Name attribute of this Accountnig-Request.
	 * 
	 * @param userName
	 *            user name to set
	 */
	public void setUserName(String userName) {
		if (userName == null)
			throw new NullPointerException("user name not set");
		if (userName.length() == 0)
			throw new IllegalArgumentException("empty user name not allowed");

		removeAttributes(USER_NAME);
		addAttribute(new StringAttribute(USER_NAME, userName));
	}

	/**
	 * Retrieves the user name from the RadUser-Name attribute.
	 * 
	 * @return user name
	 * @throws RadiusException
	 */
	public String getUserName()  {
		List attrs = getAttributes(USER_NAME);
		if (attrs.size() < 1 || attrs.size() > 1)
			return "";
//			throw new RuntimeException("exactly one RadUser-Name attribute required");

		RadiusAttribute ra = (RadiusAttribute) attrs.get(0);
		return ((StringAttribute) ra).getAttributeValue();
	}

	/**
	 * Sets the Acct-Status-Type attribute of this Accountnig-Request.
	 * 
	 * @param acctStatusType
	 *            ACCT_STATUS_TYPE_* to set
	 */
	public void setAcctStatusType(int acctStatusType) {
		if (acctStatusType < 1 || acctStatusType > 15)
			throw new IllegalArgumentException("bad Acct-Status-Type");
		removeAttributes(ACCT_STATUS_TYPE);
		addAttribute(new IntegerAttribute(ACCT_STATUS_TYPE, acctStatusType));
	}

	/**
	 * Retrieves the user name from the RadUser-Name attribute.
	 * 
	 * @return user name
	 * @throws RadiusException
	 */
	public int getAcctStatusType()  {
		RadiusAttribute ra = getAttribute(ACCT_STATUS_TYPE);
		if (ra == null) {
			return -1;
		}
		return ((IntegerAttribute) ra).getAttributeValueInt();
	}

	/**
	 * Calculates the request authenticator as specified by RFC 2866.
	 * 
	 * @see RadiusPacket#updateRequestAuthenticator(String, int, byte[])
	 */
	protected byte[] updateRequestAuthenticator(String sharedSecret, int packetLength, byte[] attributes) {
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

	/**
	 * Checks the received request authenticator as specified by RFC 2866.
	 */
	protected void checkRequestAuthenticator(String sharedSecret, int packetLength, byte[] attributes) throws RadiusException {
		byte[] expectedAuthenticator = updateRequestAuthenticator(sharedSecret, packetLength, attributes);
		byte[] receivedAuth = getAuthenticator();
		for (int i = 0; i < 16; i++)
			if (expectedAuthenticator[i] != receivedAuth[i])
				throw new RadiusException("request authenticator invalid");
	}

	/**
	 * Radius RadUser-Name attribute type
	 */
	private static final int USER_NAME = 1;

	/**
	 * Radius Acct-Status-Type attribute type
	 */
	private static final int ACCT_STATUS_TYPE = 40;

	private boolean radsec;

	public boolean isRadsec() {
		return radsec;
	}

	public void setRadsec(boolean radsec) {
		this.radsec = radsec;
	}

	/**
	 * String representation of this packet, for debugging purposes.
	 *
	 * @see Object#toString()
	 */
	public String toString() {
		StringBuffer s = new StringBuffer();
		s.append(getPacketTypeName());
		s.append("(").append(getStatusTypeName()).append(")");
		s.append(", ID ");
		s.append(getPacketIdentifier());
		for (Iterator i = getAttributes().iterator(); i.hasNext();) {
			RadiusAttribute attr = (RadiusAttribute) i.next();
			s.append("\n");
			s.append(String.format("\t%s", attr.toString()));
		}
		return s.toString();
	}

	public String toLineString() {
		StringBuffer s = new StringBuffer();
		s.append(getPacketTypeName());
		s.append("(").append(getStatusTypeName()).append(")");
		s.append(", ID ");
		s.append(getPacketIdentifier());
		for (Iterator i = getAttributes().iterator(); i.hasNext();) {
			RadiusAttribute attr = (RadiusAttribute) i.next();
			s.append(", ");
			s.append(attr.toString());
		}
		return s.toString();
	}


	public String toSimpleString() {
		StringBuffer s = new StringBuffer();
		s.append(getPacketTypeName());
		s.append("(").append(getStatusTypeName()).append(")").append(":");
		s.append(String.format("username=%s, ", getUsername()));
		s.append(String.format("macAddr=%s, ", getMacAddr()));
		s.append(String.format("nasPortId=%s, ", getNasPortId()));
		s.append(String.format("userIp=%s, ", getFramedIpaddr()));
		s.append(String.format("nasAddr=%s ", getNasAddr()));
		return s.toString();
	}

}

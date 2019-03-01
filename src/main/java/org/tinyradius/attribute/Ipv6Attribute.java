package org.tinyradius.attribute;

import org.tinyradius.util.RadiusException;

import java.net.Inet6Address;
import java.net.UnknownHostException;

/**
 * This class represents a Radius attribute for an IPv6 number.
 */
public class Ipv6Attribute extends RadiusAttribute {

	/**
	 * Constructs an empty IPv6 attribute.
	 */
	public Ipv6Attribute() {
		super();
	}
	
	/**
	 * Constructs an IPv6 attribute.
	 * @param type attribute type code
	 * @param value value, format:ipv6 address
	 */
	public Ipv6Attribute(int type, String value) {
		setAttributeType(type);
		setAttributeValue(value);
	}
	
	/**
	 * Returns the attribute value (IPv6 number) as a string of the
	 * format ipv6 address
	 * @see RadiusAttribute#getAttributeValue()
	 */
	public String getAttributeValue() {
		byte[] data = getAttributeData();
		if (data == null || data.length != 16)
			throw new RuntimeException("ip attribute: expected 16 bytes attribute data");
		try {
			Inet6Address addr = (Inet6Address)Inet6Address.getByAddress(null, data);
		
			return addr.getHostAddress();
		} catch (UnknownHostException e) {
			throw new IllegalArgumentException("bad IPv6 address", e);
		}
		
	}
	
	/**
	 * Sets the attribute value (IPv6 number). String format:
	 * ipv6 address.
	 * @throws IllegalArgumentException
	 * @throws NumberFormatException
	 * @see RadiusAttribute#setAttributeValue(String)
	 */
	public void setAttributeValue(String value) {
		if (value == null || value.length() < 3)
			throw new IllegalArgumentException("bad IPv6 address : " + value);
		try {
			final Inet6Address addr = (Inet6Address)Inet6Address.getByName(value);
		
			byte[] data = addr.getAddress();
		
			setAttributeData(data);
		} catch (UnknownHostException e) {
			throw new IllegalArgumentException("bad IPv6 address : " + value, e);
		}
	}
	

	/**
	 * Check attribute length.
	 * @see RadiusAttribute#readAttribute(byte[], int, int)
	 */
	public void readAttribute(byte[] data, int offset, int length)
	throws RadiusException {
		if (length != 18)
			throw new RadiusException("IP attribute: expected 16 bytes data");
		super.readAttribute(data, offset, length);
	}

}

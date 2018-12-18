/**
 * Created on 24/Jun/2016
 * @author Ivan F. Martinez
 */
package org.tinyradius.attribute;

import java.util.Arrays;
import java.util.StringTokenizer;
import java.net.Inet6Address;
import java.net.UnknownHostException;

import org.tinyradius.util.RadiusException;

/**
 * This class represents a Radius attribute for an IPv6 prefix.
 */
public class Ipv6PrefixAttribute extends RadiusAttribute {

	/**
	 * Constructs an empty IP attribute.
	 */
	public Ipv6PrefixAttribute() {
		super();
	}
	
	/**
	 * Constructs an IPv6 prefix attribute.
	 * @param type attribute type code
	 * @param value value, format: "ipv6 address"/prefix
	 */
	public Ipv6PrefixAttribute(int type, String value) {
		setAttributeType(type);
		setAttributeValue(value);
	}
	
	/**
	 * Returns the attribute value (IP number) as a string of the
	 * format "xx.xx.xx.xx".
	 * @see org.tinyradius.attribute.RadiusAttribute#getAttributeValue()
	 */
	public String getAttributeValue() {
		final byte[] data = getAttributeData();
		if (data == null || data.length != 18)
			throw new RuntimeException("ip attribute: expected 18 bytes attribute data");
		try {
		        final int prefix = (data[1] & 0xff);		        
			final Inet6Address addr = (Inet6Address)Inet6Address.getByAddress(null, Arrays.copyOfRange(data,2,data.length));
		
			return addr.getHostAddress() + "/" + prefix;
		} catch (UnknownHostException e) {
			throw new IllegalArgumentException("bad IPv6 prefix", e);
		}
		
	}
	
	/**
	 * Sets the attribute value (IPv6 number/prefix). String format:
	 * ipv6 address.
	 * @throws IllegalArgumentException
	 * @throws NumberFormatException
	 * @see org.tinyradius.attribute.RadiusAttribute#setAttributeValue(java.lang.String)
	 */
	public void setAttributeValue(String value) {
		if (value == null || value.length() < 3)
			throw new IllegalArgumentException("bad IPv6 address : " + value);
		try {
		        final byte[] data = new byte[18];
		        data[0] = 0;
//TODO better checking		        
		        final int slashPos = value.indexOf("/");
		        data[1] = (byte)(Integer.valueOf(value.substring(slashPos+1)) & 0xff);
		        
			final Inet6Address addr = (Inet6Address)Inet6Address.getByName(value.substring(0,slashPos));
		
			byte[] ipData = addr.getAddress();
			for (int i = 0; i < ipData.length; i++) {
			    data[i+2] = ipData[i];
			}
		
			setAttributeData(data);
		} catch (UnknownHostException e) {
			throw new IllegalArgumentException("bad IPv6 address : " + value, e);
		}
	}
	

	/**
	 * Check attribute length.
	 * @see org.tinyradius.attribute.RadiusAttribute#readAttribute(byte[], int, int)
	 */
	public void readAttribute(byte[] data, int offset, int length)
	throws RadiusException {
		if (length != 20)
			throw new RadiusException("IP attribute: expected 18 bytes data");
		super.readAttribute(data, offset, length);
	}

}

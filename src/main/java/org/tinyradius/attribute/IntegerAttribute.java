/**
 * $Id: IntegerAttribute.java,v 1.4 2005/09/04 22:11:03 wuttke Exp $
 * Created on 08.04.2005
 * @author Matthias Wuttke
 * @version $Revision: 1.4 $
 */
package org.tinyradius.attribute;

import org.tinyradius.dictionary.AttributeType;
import org.tinyradius.util.RadiusException;

/**
 * This class represents a Radius attribute which only
 * contains a 32 bit integer.
 */
public class IntegerAttribute extends RadiusAttribute {

	/**
	 * Constructs an empty integer attribute.
	 */
	public IntegerAttribute() {
		super();
	}
	
	/**
	 * Constructs an integer attribute with the given value.
	 * @param type attribute type
	 * @param value attribute value
	 */
	public IntegerAttribute(int type, int value) {
		setAttributeType(type);
		setAttributeValue(value);
	}
	
	/**
	 * Returns the string value of this attribute.
	 * @return a string
	 */
	public int getAttributeValueInt() {
		byte[] data = getAttributeData();
		return (((data[0] & 0x0ff) << 24) | ((data[1] & 0x0ff) << 16) | 
				((data[2] & 0x0ff) << 8) | (data[3] & 0x0ff));
	}
	
	/**
	 * Returns the value of this attribute as a string.
	 * Tries to resolve enumerations.
	 * @see org.tinyradius.attribute.RadiusAttribute#getAttributeValue()
	 */
	public String getAttributeValue() {
		int value = getAttributeValueInt();
		AttributeType at = getAttributeTypeObject();
		if (at != null) {
			String name = at.getEnumeration(value);
			if (name != null)
				return name;
		}
		// Radius uses only unsigned values....
		return Long.toString(((long)value & 0xffffffffl));
	}
	
	/**
	 * Sets the value of this attribute.
	 * @param value integer value
	 */
	public void setAttributeValue(int value) {
		byte[] data = new byte[4];
		data[0] = (byte)(value >> 24 & 0x0ff);
		data[1] = (byte)(value >> 16 & 0x0ff);
		data[2] = (byte)(value >> 8 & 0x0ff);
		data[3] = (byte)(value & 0x0ff);
		setAttributeData(data);
	}
	
	/**
	 * Sets the value of this attribute.
	 * @exception NumberFormatException if value is not a number and constant cannot be resolved
	 * @see org.tinyradius.attribute.RadiusAttribute#setAttributeValue(java.lang.String)
	 */
	public void setAttributeValue(String value) {
		AttributeType at = getAttributeTypeObject();
		if (at != null) {
			Integer val = at.getEnumeration(value);
			if (val != null) {
				setAttributeValue(val.intValue());
				return;
			}
		}
		
		// Radius uses only unsigned integers for this the parser should consider as Long to parse high bit correctly...
		setAttributeValue((int)Long.parseLong(value));
	}
	
	/**
	 * Check attribute length.
	 * @see org.tinyradius.attribute.RadiusAttribute#readAttribute(byte[], int, int)
	 */
	public void readAttribute(byte[] data, int offset, int length)
	throws RadiusException {
		if (length != 6)
			throw new RadiusException("integer attribute: expected 4 bytes data");
		super.readAttribute(data, offset, length);
	}
	
}

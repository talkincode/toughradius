package org.tinyradius.attribute;

import java.io.UnsupportedEncodingException;

/**
 * This class represents a Radius attribute which only
 * contains a string.
 */
public class StringAttribute extends RadiusAttribute {

	/**
	 * Constructs an empty string attribute.
	 */
	public StringAttribute() {
		super();
	}
	
	/**
	 * Constructs a string attribute with the given value.
	 * @param type attribute type
	 * @param value attribute value
	 */
	public StringAttribute(int type, String value) {
		setAttributeType(type);
		setAttributeValue(value);
	}
	
	/**
	 * Returns the string value of this attribute.
	 * @return a string
	 */
	public String getAttributeValue() {
		try {
			return new String(getAttributeData(), "UTF-8");
		} catch (UnsupportedEncodingException uee) {
			return new String(getAttributeData());
		}
	}
	
	/**
	 * Sets the string value of this attribute.
	 * @param value string, not null
	 */
	public void setAttributeValue(String value) {
		if (value == null)
			throw new NullPointerException("string value not set");
		try {
			setAttributeData(value.getBytes("UTF-8"));
		} catch (UnsupportedEncodingException uee) {
			setAttributeData(value.getBytes());
		}
	}
	
}

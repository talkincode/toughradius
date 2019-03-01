package org.tinyradius.dictionary;

import java.util.HashMap;
import java.util.Iterator;
import java.util.Map;

/**
 * A dictionary that keeps the values and names in hash maps
 * in the memory. The dictionary has to be filled using the
 * methods <code>addAttributeType</code> and <code>addVendor</code>.
 * 
 * @see #addAttributeType(AttributeType)
 * @see #addVendor(int, String)
 * @see Dictionary
 * @see WritableDictionary
 */
public class MemoryDictionary implements WritableDictionary {

	/**
	 * Returns the AttributeType for the vendor -1 from the
	 * component.
	 * 
	 * @param typeCode
	 *            attribute type code
	 * @return AttributeType or null
	 * @see Dictionary#getAttributeTypeByCode(int)
	 */
	public AttributeType getAttributeTypeByCode(int typeCode) {
		return getAttributeTypeByCode(-1, typeCode);
	}

	/**
	 * Returns the specified AttributeType object.
	 * 
	 * @param vendorCode
	 *            vendor ID or -1 for "no vendor"
	 * @param typeCode
	 *            attribute type code
	 * @return AttributeType or null
	 * @see Dictionary#getAttributeTypeByCode(int, int)
	 */
	public AttributeType getAttributeTypeByCode(int vendorCode, int typeCode) {
		Map vendorAttributes = (Map) attributesByCode.get(new Integer(vendorCode));
		if (vendorAttributes == null) {
			return null;
		}
		return (AttributeType) vendorAttributes.get(new Integer(typeCode));
	}

	/**
	 * Retrieves the attribute type with the given name.
	 * 
	 * @param typeName
	 *            name of the attribute type
	 * @return AttributeType or null
	 * @see Dictionary#getAttributeTypeByName(String)
	 */
	public AttributeType getAttributeTypeByName(String typeName) {
		return (AttributeType) attributesByName.get(typeName);
	}

	/**
	 * Searches the vendor with the given name and returns its
	 * code. This method is seldomly used.
	 *
	 * @param vendorName
	 *            vendor name
	 * @return vendor code or -1
	 * @see Dictionary#getVendorId(String)
	 */
	public int getVendorId(String vendorName) {
		for (Iterator i = vendorsByCode.entrySet().iterator(); i.hasNext();) {
			Map.Entry e = (Map.Entry) i.next();
			if (e.getValue().equals(vendorName))
				return ((Integer) e.getKey()).intValue();
		}
		return -1;
	}

	/**
	 * Retrieves the name of the vendor with the given code from
	 * the component.
	 * 
	 * @param vendorId
	 *            vendor number
	 * @return vendor name or null
	 * @see Dictionary#getVendorName(int)
	 */
	public String getVendorName(int vendorId) {
		return (String) vendorsByCode.get(new Integer(vendorId));
	}

	/**
	 * Adds the given vendor to the component.
	 * 
	 * @param vendorId
	 *            vendor ID
	 * @param vendorName
	 *            name of the vendor
	 * @exception IllegalArgumentException
	 *                empty vendor name, invalid vendor ID
	 */
	public void addVendor(int vendorId, String vendorName) {
		if (vendorId < 0)
			throw new IllegalArgumentException("vendor ID must be positive");
		if (getVendorName(vendorId) != null)
			throw new IllegalArgumentException("duplicate vendor code");
		if (vendorName == null || vendorName.length() == 0)
			throw new IllegalArgumentException("vendor name empty");
		vendorsByCode.put(new Integer(vendorId), vendorName);
	}

	/**
	 * Adds an AttributeType object to the component.
	 * 
	 * @param attributeType
	 *            AttributeType object
	 * @exception IllegalArgumentException
	 *                duplicate attribute name/type code
	 */
	public void addAttributeType(AttributeType attributeType) {
		if (attributeType == null)
			throw new IllegalArgumentException("attribute type must not be null");

		Integer vendorId = new Integer(attributeType.getVendorId());
		Integer typeCode = new Integer(attributeType.getTypeCode());
		String attributeName = attributeType.getName();

		if (attributesByName.containsKey(attributeName))
			throw new IllegalArgumentException("duplicate attribute name: " + attributeName);

		Map vendorAttributes = (Map) attributesByCode.get(vendorId);
		if (vendorAttributes == null) {
			vendorAttributes = new HashMap();
			attributesByCode.put(vendorId, vendorAttributes);
		}
		if (vendorAttributes.containsKey(typeCode))
			throw new IllegalArgumentException("duplicate type code: " + typeCode);

		attributesByName.put(attributeName, attributeType);
		vendorAttributes.put(typeCode, attributeType);
	}

	private Map vendorsByCode = new HashMap(); // <Integer, String>
	private Map attributesByCode = new HashMap(); // <Integer, <Integer, AttributeType>>
	private Map attributesByName = new HashMap(); // <String, AttributeType>

}

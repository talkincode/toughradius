package org.tinyradius.dictionary;

/**
 * A dictionary retrieves AttributeType objects by name or
 * type code. 
 */
public interface Dictionary {

	/**
	 * Retrieves an attribute type by name. This includes
	 * vendor-specific attribute types whose name is prefixed
	 * by the vendor name. 
	 * @param typeName name of the attribute type 
	 * @return AttributeType object or null
	 */
	public AttributeType getAttributeTypeByName(String typeName);
	
	/**
	 * Retrieves an attribute type by type code. This method
	 * does not retrieve vendor specific attribute types.
	 * @param typeCode type code, 1-255
	 * @return AttributeType object or null
	 */
	public AttributeType getAttributeTypeByCode(int typeCode);
	
	/**
	 * Retrieves an attribute type for a vendor-specific
	 * attribute.
	 * @param vendorId vendor ID
	 * @param typeCode type code, 1-255
	 * @return AttributeType object or null
	 */
	public AttributeType getAttributeTypeByCode(int vendorId, int typeCode);

	/**
	 * Retrieves the name of the vendor with the given
	 * vendor code.
	 * @param vendorId vendor number
	 * @return vendor name or null
	 */
	public String getVendorName(int vendorId);
	
	/**
	 * Retrieves the ID of the vendor with the given
	 * name.
	 * @param vendorName name of the vendor
	 * @return vendor ID or -1
	 */
	public int getVendorId(String vendorName);
	
}

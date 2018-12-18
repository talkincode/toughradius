/**
 * $Id: WritableDictionary.java,v 1.1 2005/09/04 22:11:00 wuttke Exp $
 * Created on 28.08.2005
 * @author mw
 * @version $Revision: 1.1 $
 */
package org.tinyradius.dictionary;

/**
 * A dictionary that is not read-only. Provides methods
 * to add entries to the dictionary.
 */
public interface WritableDictionary 
extends Dictionary {

	/**
	 * Adds the given vendor to the dictionary.
	 * @param vendorId vendor ID
	 * @param vendorName name of the vendor
	 */
	public void addVendor(int vendorId, String vendorName);

	/**
	 * Adds an AttributeType object to the dictionary.
	 * @param attributeType AttributeType object
	 */
	public void addAttributeType(AttributeType attributeType);

}

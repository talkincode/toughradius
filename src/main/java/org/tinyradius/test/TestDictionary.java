/**
 * $Id: TestDictionary.java,v 1.1 2005/09/06 16:38:41 wuttke Exp $
 * Created on 06.09.2005
 * @author mw
 * @version $Revision: 1.1 $
 */
package org.tinyradius.test;

import java.io.FileInputStream;
import java.io.InputStream;

import org.tinyradius.attribute.IpAttribute;
import org.tinyradius.attribute.Ipv6Attribute;
import org.tinyradius.attribute.Ipv6PrefixAttribute;
import org.tinyradius.dictionary.Dictionary;
import org.tinyradius.dictionary.DefaultDictionary;
import org.tinyradius.dictionary.DictionaryParser;
import org.tinyradius.packet.AccessRequest;

/**
 * Shows how to use TinyRadius with a custom dictionary
 * loaded from a dictionary file.
 * Requires a file "test.dictionary" in the current directory.
 */
public class TestDictionary {

	public static void main(String[] args) 
	throws Exception {
		Dictionary dictionary = DefaultDictionary.getDefaultDictionary();
		AccessRequest ar = new AccessRequest("UserName", "UserPassword");
		ar.setDictionary(dictionary);
		ar.addAttribute("WISPr-Location-ID", "LocationID");
		ar.addAttribute(new IpAttribute(8, 1234567));
		ar.addAttribute(new Ipv6Attribute(168, "fe80::"));
		ar.addAttribute(new Ipv6PrefixAttribute(97, "fe80::/64"));
		ar.addAttribute(new Ipv6PrefixAttribute(97, "fe80::/128"));
		System.out.println(ar);
	}
	
}

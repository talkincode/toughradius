/**
 * $Id: DictionaryParser.java,v 1.2 2005/09/06 16:38:40 wuttke Exp $
 * Created on 28.08.2005
 * 
 * @author mw
 * @version $Revision: 1.2 $
 */
package org.tinyradius.dictionary;

import java.io.BufferedReader;
import java.io.File;
import java.io.FileInputStream;
import java.io.IOException;
import java.io.InputStream;
import java.io.InputStreamReader;
import java.util.StringTokenizer;
import org.tinyradius.attribute.IntegerAttribute;
import org.tinyradius.attribute.IpAttribute;
import org.tinyradius.attribute.Ipv6Attribute;
import org.tinyradius.attribute.Ipv6PrefixAttribute;
import org.tinyradius.attribute.RadiusAttribute;
import org.tinyradius.attribute.StringAttribute;
import org.tinyradius.attribute.VendorSpecificAttribute;

/**
 * Parses a dictionary in "Radiator format" and fills a
 * WritableDictionary.
 */
public class DictionaryParser {

	/**
	 * Returns a new dictionary filled with the contents
	 * from the given input stream.
	 * 
	 * @param source
	 *            input stream
	 * @return dictionary object
	 * @throws IOException
	 */
	public static Dictionary parseDictionary(InputStream source) throws IOException {
		WritableDictionary d = new MemoryDictionary();
		parseDictionary(source, d);
		return d;
	}

	/**
	 * Parses the dictionary from the specified InputStream.
	 * 
	 * @param source
	 *            input stream
	 * @param dictionary
	 *            dictionary data is written to
	 * @throws IOException
	 *             syntax errors
	 * @throws RuntimeException
	 *             syntax errors
	 */
	public static void parseDictionary(InputStream source, WritableDictionary dictionary) throws IOException {
		// read each line separately
		BufferedReader in = new BufferedReader(new InputStreamReader(source));

		String line;
		int lineNum = -1;
		while ((line = in.readLine()) != null) {
			// ignore comments
			lineNum++;
			line = line.trim();
			if (line.startsWith("#") || line.length() == 0)
				continue;

			// tokenize line by whitespace
			StringTokenizer tok = new StringTokenizer(line);
			if (!tok.hasMoreTokens())
				continue;

			String lineType = tok.nextToken().trim();
			if (lineType.equalsIgnoreCase("ATTRIBUTE"))
				parseAttributeLine(dictionary, tok, lineNum);
			else if (lineType.equalsIgnoreCase("VALUE"))
				parseValueLine(dictionary, tok, lineNum);
			else if (lineType.equalsIgnoreCase("$INCLUDE"))
				includeDictionaryFile(dictionary, tok, lineNum);
			else if (lineType.equalsIgnoreCase("VENDORATTR"))
				parseVendorAttributeLine(dictionary, tok, lineNum);
			else if (lineType.equals("VENDOR"))
				parseVendorLine(dictionary, tok, lineNum);
			else
				throw new IOException("unknown line type: " + lineType + " line: " + lineNum);
		}
	}

	/**
	 * Parse a line that declares an attribute.
	 */
	private static void parseAttributeLine(WritableDictionary dictionary, StringTokenizer tok, int lineNum) throws IOException {
		if (tok.countTokens() != 3)
			throw new IOException("syntax error on line " + lineNum);

		// read name, code, type
		String name = tok.nextToken().trim();
		int code = Integer.parseInt(tok.nextToken());
		String typeStr = tok.nextToken().trim();

		// translate type to class
		Class type;
		if (code == VendorSpecificAttribute.VENDOR_SPECIFIC)
			type = VendorSpecificAttribute.class;
		else
			type = getAttributeTypeClass(code, typeStr);

		// create and cache object
		dictionary.addAttributeType(new AttributeType(code, name, type));
	}

	/**
	 * Parses a VALUE line containing an enumeration value.
	 */
	private static void parseValueLine(WritableDictionary dictionary, StringTokenizer tok, int lineNum) throws IOException {
		if (tok.countTokens() != 3)
			throw new IOException("syntax error on line " + lineNum);

		String typeName = tok.nextToken().trim();
		String enumName = tok.nextToken().trim();
		String valStr = tok.nextToken().trim();

		AttributeType at = dictionary.getAttributeTypeByName(typeName);
		if (at == null) {
			throw new IOException("unknown attribute type: " + typeName + ", line: " + lineNum);
		}

		at.addEnumerationValue(Integer.parseInt(valStr), enumName);
	}

	/**
	 * Parses a line that declares a Vendor-Specific attribute.
	 */
	private static void parseVendorAttributeLine(WritableDictionary dictionary, StringTokenizer tok, int lineNum) throws IOException {
		if (tok.countTokens() != 4)
			throw new IOException("syntax error on line " + lineNum);

		String vendor = tok.nextToken().trim();
		String name = tok.nextToken().trim();
		int code = Integer.parseInt(tok.nextToken().trim());
		String typeStr = tok.nextToken().trim();

		Class type = getAttributeTypeClass(code, typeStr);
		AttributeType at = new AttributeType(Integer.parseInt(vendor), code, name, type);
		dictionary.addAttributeType(at);
	}

	/**
	 * Parses a line containing a vendor declaration.
	 */
	private static void parseVendorLine(WritableDictionary dictionary, StringTokenizer tok, int lineNum) throws IOException {
		if (tok.countTokens() != 2)
			throw new IOException("syntax error on line " + lineNum);

		int vendorId = Integer.parseInt(tok.nextToken().trim());
		String vendorName = tok.nextToken().trim();

		dictionary.addVendor(vendorId, vendorName);
	}

	/**
	 * Includes a dictionary file.
	 */
	private static void includeDictionaryFile(WritableDictionary dictionary, StringTokenizer tok, int lineNum) throws IOException {
		if (tok.countTokens() != 1)
			throw new IOException("syntax error on line " + lineNum);
		String includeFile = tok.nextToken();

		File incf = new File(includeFile);
		if (!incf.exists())
			throw new IOException("inclueded file '" + includeFile + "' not found, line " + lineNum);

		FileInputStream fis = new FileInputStream(incf);
		parseDictionary(fis, dictionary);

		// line numbers begin with 0 again, but file name is
		// not mentioned in exceptions
		// furthermore, this method does not allow to include
		// classpath resources
	}

	/**
	 * Returns the RadiusAttribute descendant class for the given
	 * attribute type.
	 * 
	 * @param attributeType
	 * 
	 * @param typeStr
	 *            string|octets|integer|date|ipaddr|ipv6addr|ipv6prefix
	 * @return RadiusAttribute class or descendant
	 */
	private static Class getAttributeTypeClass(int attributeType, String typeStr) {
		Class type = RadiusAttribute.class;
		if (typeStr.equalsIgnoreCase("string"))
			type = StringAttribute.class;
		else if (typeStr.equalsIgnoreCase("octets"))
			type = RadiusAttribute.class;
		else if (typeStr.equalsIgnoreCase("integer") || typeStr.equalsIgnoreCase("date"))
			type = IntegerAttribute.class;
		else if (typeStr.equalsIgnoreCase("ipaddr"))
			type = IpAttribute.class;
		else if (typeStr.equalsIgnoreCase("ipv6addr"))
			type = Ipv6Attribute.class;
		else if (typeStr.equalsIgnoreCase("ipv6prefix"))
			type = Ipv6PrefixAttribute.class;
		return type;
	}

}

package org.tinyradius.attribute;

import static org.junit.Assert.*;

import org.junit.Test;

public class IntegerAttributeTest {

	@Test
	public void test() {
		final IntegerAttribute intAttr = new IntegerAttribute(27, 0);
		final long bigValue = 0xffffffffl; // big value with highest bit set
		System.err.println((int)bigValue);
		System.err.println(bigValue);
		final String bigValueSt = Long.toString(bigValue);
		intAttr.setAttributeValue(bigValueSt);
		assertEquals(bigValueSt, intAttr.getAttributeValue());
	}

}

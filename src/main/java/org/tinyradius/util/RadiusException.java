package org.tinyradius.util;

/**
 * An exception which occurs on Radius protocol errors like
 * invalid packets or malformed attributes.
 */
public class RadiusException extends Exception {

	/**
	 * Constructs a RadiusException with a message.
	 * @param message error message
	 */
	public RadiusException(String message) {
		super(message);
	}

	private static final long serialVersionUID = 2201204523946051388L;

}

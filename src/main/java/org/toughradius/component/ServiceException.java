package org.toughradius.component;

/**
 * An exception which occurs on Radius protocol errors like
 * invalid packets or malformed attributes.
 */
public class ServiceException extends Exception {

	/**
	 * Constructs a RadiusException with a message.
	 * @param message error message
	 */
	public ServiceException(String message) {
		super(message);
	}

	public ServiceException(String message, Throwable cause) {
		super(message, cause);
	}

	private static final long serialVersionUID = 2201204523946051388L;

}

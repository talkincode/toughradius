package org.toughradius.component;


public class ServiceException extends Exception {

	public ServiceException(String message) {
		super(message);
	}

	public ServiceException(String message, Throwable cause) {
		super(message, cause);
	}

	private static final long serialVersionUID = 2201204523946051388L;

}

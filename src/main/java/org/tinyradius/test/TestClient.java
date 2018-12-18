/**
 * $Id: TestClient.java,v 1.4 2006/02/17 18:14:54 wuttke Exp $
 * Created on 08.04.2005
 * @author Matthias Wuttke
 * @version $Revision: 1.4 $
 */
package org.tinyradius.test;

import org.tinyradius.packet.AccessRequest;
import org.tinyradius.packet.AccountingRequest;
import org.tinyradius.packet.RadiusPacket;
import org.tinyradius.util.RadiusClient;

/**
 * Simple Radius command-line client.
 */
public class TestClient {
	
	/**
	 * Radius command line client.
	 * <br/>Usage: TestClient <i>hostName sharedSecret userName password</i>
	 * @param args arguments
	 * @throws Exception
	 */
	public static void main(String[] args) 
	throws Exception {
		if (args.length != 4) {
			System.out.println("Usage: TestClient hostName sharedSecret userName password");
			System.exit(1);
		}
		
		String host = args[0];
		String shared = args[1];
		String user = args[2];
		String pass = args[3];
		
		RadiusClient rc = new RadiusClient(host, shared);

		// 1. Send Access-Request
		AccessRequest ar = new AccessRequest(user, pass);
		ar.setAuthProtocol(AccessRequest.AUTH_PAP); // or AUTH_CHAP
		ar.addAttribute("NAS-Identifier", "this.is.my.nas-identifier.de");
		ar.addAttribute("NAS-IP-Address", "192.168.0.100");
		ar.addAttribute("Service-Type", "Login-User");
		ar.addAttribute("WISPr-Redirection-URL", "http://www.sourceforge.net/");
		ar.addAttribute("WISPr-Location-ID", "net.sourceforge.ap1");
		
		System.out.println("Packet before it is sent\n" + ar + "\n");
		RadiusPacket response = rc.authenticate(ar);
		System.out.println("Packet after it was sent\n" + ar + "\n");
		System.out.println("Response\n" + response + "\n");

		// 2. Send Accounting-Request
		AccountingRequest acc = new AccountingRequest("mw", AccountingRequest.ACCT_STATUS_TYPE_START);
		acc.addAttribute("Acct-Session-Id", "1234567890");
		acc.addAttribute("NAS-Identifier", "this.is.my.nas-identifier.de");
		acc.addAttribute("NAS-Port", "0");
	
		System.out.println(acc + "\n");	
		response = rc.account(acc);
		System.out.println("Response: " + response);
		
		rc.close();
	}
	
}

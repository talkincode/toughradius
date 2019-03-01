package org.tinyradius.util;

import gnu.crypto.hash.HashFactory;
import gnu.crypto.hash.IMessageDigest;
import net.sf.jradius.util.MSCHAP;
import net.sf.jradius.util.RadiusUtils;

/**
 *
 * @see //http://www.ietf.org/rfc/rfc2759.txt
 */
public class Authenticator {

	private static final byte[] Magic1 = new byte[]{
		0x4D, 0x61, 0x67, 0x69, 0x63, 0x20, 0x73, 0x65, 0x72, 0x76,
		0x65, 0x72, 0x20, 0x74, 0x6F, 0x20, 0x63, 0x6C, 0x69, 0x65,
		0x6E, 0x74, 0x20, 0x73, 0x69, 0x67, 0x6E, 0x69, 0x6E, 0x67,
		0x20, 0x63, 0x6F, 0x6E, 0x73, 0x74, 0x61, 0x6E, 0x74};
	
	
//	{0x4D, 0x61, 0x67, 0x69, 0x63, 0x20, 0x73, 0x65, 0x72, 0x76, 
//		0x65, 0x72, 0x20, 0x74, 0x6F, 0x20, 0x63, 0x6C, 0x69, 0x65,
//		0x6E, 0x74, 0x20, 0x73, 0x69, 0x67, 0x6E, 0x69, 0x6E, 0x67, 
//		0x20, 0x63, 0x6F, 0x6E, 0x73, 0x74, 0x61, 0x6E, 0x74, }

	private static final byte[] Magic2 = new byte[]{
		0x50, 0x61, 0x64, 0x20, 0x74, 0x6F, 0x20, 0x6D, 0x61, 0x6B,
		0x65, 0x20, 0x69, 0x74, 0x20, 0x64, 0x6F, 0x20, 0x6D, 0x6F,
		0x72, 0x65, 0x20, 0x74, 0x68, 0x61, 0x6E, 0x20, 0x6F, 0x6E,
		0x65, 0x20, 0x69, 0x74, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6F,
		0x6E};

	public static byte[] GenerateAuthenticatorResponse(
			byte[] password,
			byte[] ntResponse,
			byte[] peerChallenge,
			byte[] authenticatorChallenge,
			byte[] userName) {
		if(password == null) {
			password = new byte[] { 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00 }; 
		}
		byte[] passwordHashHash = getPasswordHashHash(password);

		IMessageDigest md = HashFactory.getInstance("SHA-1");
		md.update(passwordHashHash, 0, 16);
		md.update(ntResponse, 0, 24);
		md.update(Magic1, 0, 39);
		byte[] digest = md.digest();
//		System.err.println("digest (" + RadiusUtils.byteArrayToHexString(digest) + ")");
//		System.err.println("Magic1 (" + RadiusUtils.byteArrayToHexString(Magic1) + ")");

		byte[] challenge = MSCHAP.ChallengeHash(peerChallenge, authenticatorChallenge, userName);
//		System.err.println("challenge (" + RadiusUtils.byteArrayToHexString(challenge) + ")");

		IMessageDigest md2 = HashFactory.getInstance("SHA-1");
		md2.update(digest, 0, 20);
		md2.update(challenge, 0, 8);
		md2.update(Magic2, 0, 41);

		byte[] authenticatorResponse = md2.digest();
//		System.err.println("digest (" + RadiusUtils.byteArrayToHexString(authenticatorResponse) + ")");

		return authenticatorResponse;
	}

	public static byte[] getPasswordHashHash(byte[] password) {
//		System.err.println("password (" + new String(password) + ")");
		byte[] passwordHash = MSCHAP.NtPasswordHash(password);
//		System.err.println("passwordHash (" + RadiusUtils.byteArrayToHexString(passwordHash) + ")");
		byte[] passwordHashHash = MSCHAP.HashNtPasswordHash(passwordHash);
//		System.err.println("passwordHashHash (" + RadiusUtils.byteArrayToHexString(passwordHashHash) + ")");
		return passwordHashHash;
	}

	public static void main(String[] args) {
		String username = "User";
		String password = "clientPass";

		byte[] ntResponse = new byte[] {
			(byte) 0x82, (byte) 0x30, (byte) 0x9E, (byte) 0xCD, (byte) 0x8D,
			(byte) 0x70, (byte) 0x8B, (byte) 0x5E, (byte) 0xA0, (byte) 0x8F,
			(byte) 0xAA, (byte) 0x39, (byte) 0x81, (byte) 0xCD, (byte) 0x83,
			(byte) 0x54, (byte) 0x42, (byte) 0x33, (byte) 0x11, (byte) 0x4A,
			(byte) 0x3D, (byte) 0x85, (byte) 0xD6, (byte) 0xDF
		};

		byte[] peerChallenge = new byte[] {
			0x21, 0x40, 0x23, 0x24, 0x25,
			0x5E, 0x26, 0x2A, 0x28, 0x29,
			0x5F, 0x2B, 0x3A, 0x33, 0x7C,
			0x7E
		};

		byte[] authenticatorChallenge = new byte[] {
			0x5B, 0x5D, 0x7C, 0x7D, 0x7B,
			0x3F, 0x2F, 0x3E, 0x3C, 0x2C,
			0x60, 0x21, 0x32, 0x26, 0x26,
			0x28
		};

		byte[] authResponse = GenerateAuthenticatorResponse(password.getBytes(),
				ntResponse,
				peerChallenge,
				authenticatorChallenge,
				username.getBytes());
		System.out.println("authResponse: " + RadiusUtils.byteArrayToHexString(authResponse));
		//authResponse: 407a5589115fd0d6209f510fe9c04566932cda56
	}
}
package org.tinyradius.util;

import org.toughradius.common.CoderUtil;
import org.toughradius.common.bits.NetBits;
import gnu.crypto.hash.HashFactory;
import gnu.crypto.hash.IMessageDigest;

import java.io.UnsupportedEncodingException;
import java.net.Inet4Address;
import java.net.UnknownHostException;
import java.security.MessageDigest;
import java.security.NoSuchAlgorithmException;
import java.security.SecureRandom;

/**
 * This class contains miscellaneous static utility functions.
 */
public class RadiusUtil {


	static byte magic1[] = new byte[] { 0x54, 0x68, 0x69, 0x73, 0x20, 0x69,
			0x73, 0x20, 0x74, 0x68, 0x65, 0x20, 0x4d, 0x50, 0x50, 0x45, 0x20,
			0x4d, 0x61, 0x73, 0x74, 0x65, 0x72, 0x20, 0x4b, 0x65, 0x79 };

	static byte magic2[] = new byte[] { 0x4f, 0x6e, 0x20, 0x74, 0x68, 0x65,
			0x20, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x20, 0x73, 0x69, 0x64,
			0x65, 0x2c, 0x20, 0x74, 0x68, 0x69, 0x73, 0x20, 0x69, 0x73, 0x20,
			0x74, 0x68, 0x65, 0x20, 0x73, 0x65, 0x6e, 0x64, 0x20, 0x6b, 0x65,
			0x79, 0x3b, 0x20, 0x6f, 0x6e, 0x20, 0x74, 0x68, 0x65, 0x20, 0x73,
			0x65, 0x72, 0x76, 0x65, 0x72, 0x20, 0x73, 0x69, 0x64, 0x65, 0x2c,
			0x20, 0x69, 0x74, 0x20, 0x69, 0x73, 0x20, 0x74, 0x68, 0x65, 0x20,
			0x72, 0x65, 0x63, 0x65, 0x69, 0x76, 0x65, 0x20, 0x6b, 0x65, 0x79,
			0x2e };

	static byte magic3[] = new byte[] { 0x4f, 0x6e, 0x20, 0x74, 0x68, 0x65,
			0x20, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x20, 0x73, 0x69, 0x64,
			0x65, 0x2c, 0x20, 0x74, 0x68, 0x69, 0x73, 0x20, 0x69, 0x73, 0x20,
			0x74, 0x68, 0x65, 0x20, 0x72, 0x65, 0x63, 0x65, 0x69, 0x76, 0x65,
			0x20, 0x6b, 0x65, 0x79, 0x3b, 0x20, 0x6f, 0x6e, 0x20, 0x74, 0x68,
			0x65, 0x20, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x20, 0x73, 0x69,
			0x64, 0x65, 0x2c, 0x20, 0x69, 0x74, 0x20, 0x69, 0x73, 0x20, 0x74,
			0x68, 0x65, 0x20, 0x73, 0x65, 0x6e, 0x64, 0x20, 0x6b, 0x65, 0x79,
			0x2e };

	static byte SHSpad1[] = { 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00 };

	static byte SHSpad2[] = { (byte) 0xf2, (byte) 0xf2, (byte) 0xf2,
			(byte) 0xf2, (byte) 0xf2, (byte) 0xf2, (byte) 0xf2, (byte) 0xf2,
			(byte) 0xf2, (byte) 0xf2, (byte) 0xf2, (byte) 0xf2, (byte) 0xf2,
			(byte) 0xf2, (byte) 0xf2, (byte) 0xf2, (byte) 0xf2, (byte) 0xf2,
			(byte) 0xf2, (byte) 0xf2, (byte) 0xf2, (byte) 0xf2, (byte) 0xf2,
			(byte) 0xf2, (byte) 0xf2, (byte) 0xf2, (byte) 0xf2, (byte) 0xf2,
			(byte) 0xf2, (byte) 0xf2, (byte) 0xf2, (byte) 0xf2, (byte) 0xf2,
			(byte) 0xf2, (byte) 0xf2, (byte) 0xf2, (byte) 0xf2, (byte) 0xf2,
			(byte) 0xf2, (byte) 0xf2 };

	/**
	 * Random number generator.
	 */
	private static SecureRandom random = new SecureRandom();
	/**
	 * Returns the passed string as a byte array containing the
	 * string in UTF-8 representation.
	 * @param str Java string
	 * @return UTF-8 byte array
	 */
	public static byte[] getUtf8Bytes(String str) {
		try {
			return str.getBytes("UTF-8");
		} catch (UnsupportedEncodingException uee) {
			return str.getBytes();
		}
	}
	
	/**
	 * Creates a string from the passed byte array containing the
	 * string in UTF-8 representation.
	 * @param utf8 UTF-8 byte array
	 * @return Java string
	 */
	public static String getStringFromUtf8(byte[] utf8) {
		try {
			return new String(utf8, "UTF-8");
		} catch (UnsupportedEncodingException uee) {
			return new String(utf8);
		}
	}

	public static byte[] encodeString(String str) {
		try {
			return str.getBytes("UTF-8");
		} catch (UnsupportedEncodingException uee) {
			return str.getBytes();
		}
	}


	public static String decodeString(byte[] utf8) {
		try {
			return new String(utf8, "UTF-8");
		} catch (UnsupportedEncodingException uee) {
			return new String(utf8);
		}
	}

	public static byte[] encodeShort(short val){
		byte[] b = new byte[2];
		b[1] = (byte) (val >>> 0);
		b[0] = (byte) (val >>> 8);
		return b;
	}

	public static short decodeShort(byte[] b){
		return (short) (((b[1] & 0xFF) << 0) + ((b[0] & 0xFF) << 8));
	}

	public static byte[] encodeInt(int val){
		byte[] b = new byte[4];
		b[3] = (byte) (val >>> 0);
		b[2] = (byte) (val >>> 8);
		b[1] = (byte) (val >>> 16);
		b[0] = (byte) (val >>> 24);
		return b;
	}

	public static int decodeInt(byte b[]){
		return ((b[3] & 0xFF) << 0) + ((b[2] & 0xFF) << 8)
				+ ((b[1] & 0xFF) << 16) + ((b[0] & 0xFF) << 24);
	}


	public static String getHexString(byte[] data) {
		StringBuffer hex = new StringBuffer("0x");
		if (data != null)
			for (int i = 0; i < data.length; i++) {
				String digit = Integer.toString(data[i] & 0x0ff, 16);
				if (digit.length() < 2)
					hex.append('0');
				hex.append(digit);
			}
		return hex.toString();
	}

	public static String decodeIpv4(byte[] src){
		if (src.length!=4)
			throw new IllegalArgumentException("bad IP bytes");
		return (src[0] & 0xff) + "." + (src[1] & 0xff) + "." + (src[2] & 0xff) + "." + (src[3] & 0xff);
	}

	public static byte[] encodeIpV4(String value){
		try {
			return Inet4Address.getByName(value).getAddress();
		} catch (UnknownHostException e) {
			throw new IllegalArgumentException("bad IP number");
		}
	}

	public static byte[] encodeMacAddr(String value){
		if (value == null || value.length() != 17)
			throw new IllegalArgumentException("bad mac");

		value = value.replaceAll("-",":");
		byte []macBytes = new byte[6];
		String [] strArr = value.split(":");

		for(int i = 0;i < strArr.length; i++){
			int val = Integer.parseInt(strArr[i],16);
			macBytes[i] = (byte) val;
		}
		return macBytes;
	}

	public static String decodeMacAddr(byte [] src){
		String value = "";
		for(int i = 0;i < src.length; i++){
			String sTemp = Integer.toHexString(0xFF &  src[i]);
			if(sTemp.equals("0")){
				sTemp += "0";
			}
			value = value+sTemp+":";
		}
		return value.substring(0,value.lastIndexOf(":"));
	}


	/** PAP加密 */
	public static byte[] papEncryption(String userPassword, String secret, byte[] authenticator)
	{
		byte[] buf = new byte[16 + NetBits.getByteLen(secret)];
		NetBits.putString(buf, 0, secret);
		NetBits.putBytes(buf, NetBits.getByteLen(secret), authenticator);
		byte[] md5buf = CoderUtil.md5EncoderByte(buf);

		byte[] src = userPassword.getBytes();
		int byteLen = src.length>16?src.length:16;//取大
		int xorLen = src.length>16?16:src.length;//取小
		byte[] enpassword = new byte[byteLen];

		for (int i=0;i<xorLen;i++)
		{
			enpassword[i] = (byte)(src[i] ^ md5buf[i]);
		}

		if (src.length > 16)
			System.arraycopy(src, 16, enpassword, 16, src.length-16);
		else
			System.arraycopy(md5buf, src.length, enpassword, src.length, 16-src.length);

		return enpassword;
	}

	/** CHAP 加密 */
	public static byte[] chapEncryption(String userPassword, int chapId, byte[] challenge)
	{//Secret chapPassword = MD5（Chap ID + userPassword + challenge）
		byte[] buf = new byte[1 + NetBits.getByteLen(userPassword) + challenge.length];
		NetBits.putByte(buf, 0, (byte)chapId);//Chap ID
		NetBits.putString(buf, 1, userPassword);//Password
		NetBits.putBytes(buf, 1+ NetBits.getByteLen(userPassword), challenge);
		byte[] md5buf = CoderUtil.md5EncoderByte(buf);

		return md5buf;
	}

	/** PAP认证 */
	public static boolean isValidPAP(String userPassword, String secret, byte[] authenticator, byte[] userPassword2)
	{
		byte[] enPassword = papEncryption(userPassword, secret, authenticator);

		if (enPassword.length != userPassword2.length)
			return false;

		for (int i=0;i<enPassword.length;i++)
		{
			if (enPassword[i] != userPassword2[i])
				return false;
		}

		return true;
	}

	/** CHAP认证 */
	public static boolean isValidCHAP(String userPassword, int chapId, byte[] challenge, byte[] chapPassword)
	{//Secret chapPassword = MD5（Chap ID + userPassword + challenge）
		byte[] md5buf = chapEncryption(userPassword, chapId, challenge);

		if (md5buf.length != chapPassword.length)
			return false;

		for (int i=0;i<md5buf.length;i++)
		{
			if (md5buf[i] != chapPassword[i])
				return false;
		}

		return true;
	}


	/**
	 * Concatenate two byte arrays
	 * @param a
	 * @param b
	 * @return
	 */
	public static byte[] concatenateByteArrays(byte a[], byte b[]) {
		byte rv[] = new byte[a.length + b.length];

		System.arraycopy(a, 0, rv, 0, a.length);
		System.arraycopy(b, 0, rv, a.length, b.length);

		return rv;
	}

	/**
	 * Generate the MPPE Master key
	 * @see //https://tools.ietf.org/html/rfc3079#section-3
	 * @param ntHashHash
	 * @param ntResponse
	 * @return
	 */
	public static byte[] generateMPPEMasterKey(byte[] ntHashHash,
											   byte[] ntResponse) {
		IMessageDigest md = HashFactory.getInstance("SHA-1");

		md.update(ntHashHash, 0, ntHashHash.length);
		md.update(ntResponse, 0, ntResponse.length);
		md.update(magic1, 0, magic1.length);

		byte[] digest = md.digest();

		byte[] rv = new byte[16];
		System.arraycopy(digest, 0, rv, 0, 16);

		return rv;
	}

	/**
	 * Generate the MPPE Asymmetric start key
	 * @see //https://tools.ietf.org/html/rfc3079#section-3
	 * @param masterKey
	 * @param keyLength
	 * @param isSend
	 * @return
	 */
	public static byte[] generateMPPEAssymetricStartKey(byte[] masterKey,
														int keyLength, boolean isSend) {
		byte[] magic = (isSend) ? magic3 : magic2;

		IMessageDigest md = HashFactory.getInstance("SHA-1");

		md.update(masterKey, 0, 16);
		md.update(SHSpad1, 0, 40);
		md.update(magic, 0, 84);
		md.update(SHSpad2, 0, 40);

		byte[] digest = md.digest();

		byte[] rv = new byte[keyLength];
		System.arraycopy(digest, 0, rv, 0, keyLength);

		return rv;
	}

	/**
	 * Generate the MPPE Send Key (Server)
	 * @see //https://tools.ietf.org/html/rfc3079#section-3
	 * @param ntHashHash
	 * @param ntResponse
	 * @return
	 */
	public static byte[] mppeCHAP2GenKeySend128(byte[] ntHashHash, byte[] ntResponse) {
		byte[] masterKey = generateMPPEMasterKey(ntHashHash, ntResponse);

		return generateMPPEAssymetricStartKey(masterKey, 16, true);
	}

	/**
	 * Generate the MPPE Receive Key (Server)
	 * @see //https://tools.ietf.org/html/rfc3079#section-3
	 * @param ntHashHash
	 * @param ntResponse
	 * @return
	 */
	public static byte[] mppeCHAP2GenKeyRecv128(byte[] ntHashHash,
												byte[] ntResponse) {
		byte[] masterKey = generateMPPEMasterKey(ntHashHash, ntResponse);

		return generateMPPEAssymetricStartKey(masterKey, 16, false);
	}

	/**
	 * Encrypt an MPPE password
	 * Adapted from FreeRadius src/lib/radius.x:make_tunnel_password
	 * @see ??
	 * @param input
	 * 			the data to encrypt
	 * @param room
	 * 			not sure - just set it to something greater than 255
	 * @param secret
	 * 			the Radius secret for this packet
	 * @param vector
	 * 			the auth challenge
	 * @return
	 */
	public static byte[] generateEncryptedMPPEPassword(byte[] input, int room, byte[] secret, byte[] vector) {
		final int authVectorLength = 16;
		final int authPasswordLength = authVectorLength;
		final int maxStringLength = 254;

		// NOTE This could be dodgy!
		int saltOffset = 0;

		// byte digest[] = new byte[authVectorLength];
		byte passwd[] = new byte[maxStringLength + authVectorLength];
		int len;

		/*
		 * Be paranoid.
		 */
		if (room > 253)
			room = 253;

		/*
		 * Account for 2 bytes of the salt, and round the room available down to
		 * the nearest multiple of 16. Then, subtract one from that to account
		 * for the length byte, and the resulting number is the upper bound on
		 * the data to copy.
		 *
		 * We could short-cut this calculation just be forcing inlen to be no
		 * more than 239. It would work for all VSA's, as we don't pack multiple
		 * VSA's into one attribute.
		 *
		 * However, this calculation is more general, if a little complex. And
		 * it will work in the future for all possible kinds of weird attribute
		 * packing.
		 */
		room -= 2;
		room -= (room & 0x0f);
		room--;

		int inlen = input.length;

		if (inlen > room)
			inlen = room;

		/*
		 * Length of the encrypted data is password length plus one byte for the
		 * length of the password.
		 */
		len = inlen + 1;
		if ((len & 0x0f) != 0) {
			len += 0x0f;
			len &= ~0x0f;
		}

		/*
		 * Copy the password over.
		 */
		System.arraycopy(input, 0, passwd, 3, inlen);
		// memcpy(passwd + 3, input, inlen);
		for (int i = 3 + inlen; i < passwd.length - 3 - inlen; i++) {
			passwd[i] = 0;
		}
		// memset(passwd + 3 + inlen, 0, passwd.length - 3 - inlen);

		/*
		 * Generate salt. The RFC's say:
		 *
		 * The high bit of salt[0] must be set, each salt in a packet should be
		 * unique, and they should be random
		 *
		 * So, we set the high bit, add in a counter, and then add in some
		 * CSPRNG data. should be OK..
		 */
		passwd[0] = (byte) (0x80 | (((saltOffset++) & 0x0f) << 3) | (random.generateSeed(1)[0] & 0x07));
		passwd[1] = random.generateSeed(1)[0];
		passwd[2] = (byte) inlen; /* length of the password string */

		MessageDigest md5Digest = null;
		MessageDigest originalDigest = null;

		MessageDigest currentDigest = null;

		try {
			md5Digest = MessageDigest.getInstance("MD5");
			originalDigest = MessageDigest.getInstance("MD5");
		} catch (NoSuchAlgorithmException nsae) {
			throw new RuntimeException("md5 digest not available", nsae);
		}

		md5Digest.update(secret);
		originalDigest.update(secret);

		currentDigest = md5Digest;

		md5Digest.update(vector, 0, authVectorLength);
		md5Digest.update(passwd, 0, 2);

		for (int n = 0; n < len; n += authPasswordLength) {
			if (n > 0) {
				currentDigest = originalDigest;

				currentDigest.update(passwd, 2 + n - authPasswordLength, authPasswordLength);
			}

			byte digest[] = currentDigest.digest();

			for (int i = 0; i < authPasswordLength; i++) {
				passwd[i + 2 + n] ^= digest[i];
			}
		}
		byte output[] = new byte[len + 2];
		System.arraycopy(passwd, 0, output, 0, len + 2);

		return output;
	}

	protected MessageDigest getMd5Digest() {
		MessageDigest md5Digest = null;

		try {
			md5Digest = MessageDigest.getInstance("MD5");
		} catch (NoSuchAlgorithmException nsae) {
			throw new RuntimeException("md5 digest not available", nsae);
		}
		return md5Digest;
	}












}

/**
 * JRadius - A RADIUS Server Java Adapter
 * Copyright (C) 2004-2005 PicoPoint, B.V.
 *
 * This library is free software; you can redistribute it and/or modify it
 * under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation; either version 2.1 of the License, or (at
 * your option) any later version.
 *
 * This library is distributed in the hope that it will be useful, but
 * WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY
 * or FITNESS FOR A PARTICULAR PURPOSE. See the GNU Lesser General Public
 * License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with this library; if not, write to the Free Software Foundation,
 * Inc., 59 Temple Place, Suite 330, Boston, MA 02111-1307 USA
 *
 */

package net.sf.jradius.util;

import java.security.MessageDigest;
import java.util.Random;

/**
 * Radius Utilities
 * 
 * @author David Bird
 */
public final class RadiusUtils
{
    /*
     * The make*Authenticator and encodePapPassword functions are 
     * borrowed from jradius-client, also a sourceforge project 
     * and under GPL.
     */
    /**
     * This method encodes the plaintext user password according to RFC 2865
     * @param userPass java.lang.String the password to encrypt
     * @param requestAuthenticator byte[] the requestAuthenicator to use in the encryption
     * @return byte[] the byte array containing the encrypted password
     */
    public static byte[] encodePapPassword(
            final MessageDigest md5,
            final byte[] userPass, 
            final byte[] requestAuthenticator,
            final String sharedSecret) 
    {
        // encrypt the password.
        byte[] userPassBytes = null;
        //the password must be a multiple of 16 bytes and less than or equal
        //to 128 bytes. If it isn't a multiple of 16 bytes fill it out with zeroes
        //to make it a multiple of 16 bytes. If it is greater than 128 bytes
        //truncate it at 128
 
        if (userPass.length > 128)
        {
            userPassBytes = new byte[128];
            System.arraycopy(userPass,0,userPassBytes,0,128);
        }
        else 
        {
            userPassBytes = userPass;
        }
        
        // declare the byte array to hold the final product
        byte[] encryptedPass = null;
 
        if (userPassBytes.length < 128) 
        {
            if (userPassBytes.length % 16 == 0) 
            {
                // It is already a multiple of 16 bytes
                encryptedPass = new byte[userPassBytes.length];
            } 
            else 
            {
                // Make it a multiple of 16 bytes
                encryptedPass = new byte[((userPassBytes.length / 16) * 16) + 16];
            }
        } 
        else 
        {
            // the encrypted password must be between 16 and 128 bytes
            encryptedPass = new byte[128];
        }
                                                                                                                  
        // copy the userPass into the encrypted pass and then fill it out with zeroes
        System.arraycopy(userPassBytes, 0, encryptedPass, 0, userPassBytes.length);
        for(int i = userPassBytes.length; i < encryptedPass.length; i++) 
        {
            encryptedPass[i] = 0;  //fill it out with zeroes
        }
        md5.reset();
        // add the shared secret
        md5.update(sharedSecret.getBytes());
        // add the  Request Authenticator.
        md5.update(requestAuthenticator);
        // get the md5 hash( b1 = MD5(S + RA) ).
        byte bn[] = md5.digest();
                                                                                                                  
        for (int i = 0; i < 16; i++)
        {
            // perform the XOR as specified by RFC 2865.
            encryptedPass[i] = (byte)(bn[i] ^ encryptedPass[i]);
        }
                                                                                                                  
        if (encryptedPass.length > 16)
        {
            for (int i = 16; i < encryptedPass.length; i+=16)
            {
                md5.reset();
                // add the shared secret
                md5.update(sharedSecret.getBytes());
                //add the previous(encrypted) 16 bytes of the user password
                md5.update(encryptedPass, i - 16, 16);
                // get the md5 hash( bn = MD5(S + c(i-1)) ).
                bn = md5.digest();
                for (int j = 0; j < 16; j++) 
                {
                    // perform the XOR as specified by RFC 2865.
                    encryptedPass[i+j] = (byte)(bn[j] ^ encryptedPass[i+j]);
                }
            }
        }
        
        return encryptedPass;
    }
                                                                                                                 
    /**
     * This method builds a Request Authenticator for use in outgoing RADIUS
     * Access-Request packets as specified in RFC 2865.
     * @return byte[]
     */
    public static byte[] makeRFC2865RequestAuthenticator(
            final MessageDigest md5,
            final String sharedSecret) 
    {
        byte [] requestAuthenticator = new byte [16];
 
        Random r = new Random();
 
        for (int i = 0; i < 16; i++)
        {
            requestAuthenticator[i] = (byte) r.nextInt();
        }
        md5.reset();
        md5.update(sharedSecret.getBytes());
        md5.update(requestAuthenticator);
 
        return md5.digest();
    }

    /**
     * This method builds a Response Authenticator for use in validating
     * responses from the RADIUS Authentication process as specified in RFC 2865.
     * The byte array returned should match exactly the response authenticator
     * recieved in the response packet.
     * @param code byte
     * @param identifier byte
     * @param length short
     * @param requestAuthenticator byte[]
     * @param responseAttributeBytes byte[]
     * @return byte[]
     */
    public static byte[] makeRFC2865ResponseAuthenticator(
            final MessageDigest md5,
            final String sharedSecret,
            byte code,
            byte identifier,
            short length,
            byte[] requestAuthenticator,
            byte[] responseAttributeBytes) 
    {
        md5.reset();
                                                                                                                  
        md5.update((byte)code);
        md5.update((byte)identifier);
        md5.update((byte)(length >> 8));
        md5.update((byte)(length & 0xff));
        md5.update(requestAuthenticator, 0, requestAuthenticator.length);
        md5.update(responseAttributeBytes, 0, responseAttributeBytes.length);
        md5.update(sharedSecret.getBytes());
                                                                                                                  
        return md5.digest();
    }

    /**
     * This method builds a Request Authenticator for use in RADIUS Accounting
     * packets as specified in RFC 2866.
     * @param code byte
     * @param identifier byte
     * @param length short
     * @param requestAttributes byte[]
     * @return byte[]
     */
    public static byte[] makeRFC2866RequestAuthenticator(
            final MessageDigest md5,
            final String sharedSecret,
            byte code,
            byte identifier,
            int length,
            byte[] requestAttributes) 
    {
        byte [] requestAuthenticator = new byte [16];
                                                                                                                  
        for (int i = 0; i < 16; i++) 
        {
                requestAuthenticator[i] = 0;
        }
        md5.reset();
                                                                                                                  
        md5.update((byte)code);
        md5.update((byte)identifier);
        md5.update((byte)(length >> 8));
        md5.update((byte)(length & 0xff));
        md5.update(requestAuthenticator, 0, requestAuthenticator.length);
        md5.update(requestAttributes, 0, requestAttributes.length);
        md5.update(sharedSecret.getBytes());
                                                                                                                  
        return md5.digest();
    }
    
    /**
     * Converts a hex array to a human readable string
     * @param in bytes to be hexed
     * @return Returns a hex string
     */
    public static String byteArrayToHexString(byte in[])
    {
        byte ch = 0x00;
        int i = 0;
        if (in == null || in.length <= 0) return null;
        String pseudo[] = { "0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "a", "b", "c", "d", "e", "f" };
        StringBuffer out = new StringBuffer(in.length * 2);
        while (i < in.length)
        {
            ch = (byte) (in[i] & 0xF0);
            ch = (byte) (ch >>> 4);
            ch = (byte) (ch & 0x0F);
            out.append(pseudo[(int) ch]);
            ch = (byte) (in[i] & 0x0F);
            out.append(pseudo[(int) ch]);
            i++;
        }
        String rslt = new String(out);
        return rslt;
    }
}
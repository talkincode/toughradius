package org.toughradius.portal.utils;

import org.toughradius.common.CoderUtil;
import org.toughradius.common.bits.NetBits;

import java.io.UnsupportedEncodingException;
import java.net.Inet4Address;
import java.net.UnknownHostException;

public class PortalUtils {

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


}

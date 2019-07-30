package org.toughradius.common;

import org.toughradius.common.bits.NetBits;
import org.toughradius.common.coder.Base64;
import org.toughradius.common.coder.DES;
import org.toughradius.common.coder.UUID;

import java.io.UnsupportedEncodingException;
import java.net.URLEncoder;
import java.security.MessageDigest;
import java.util.Random;

/**
 * 编码工具类,包括: <br>
 * 1)MD5/SHA编码 <br>
 * 2)Base64编码 <br>
 * 3)DES/ThreeDES编码 <br>
 * 4)GBK/UTF-8/ISO-8859-1/UNICODE转换 <br>
 */
public class CoderUtil 
{
	public static final String MD5_SALT = "TOUGHRADIUS-!@#$%^";
	public static final String ALGORITHM_MD5 = "MD5";
	public static final String ALGORITHM_SHA = "SHA";
	public static final String ALGORITHM_DES = "DES";
	public static final String ALGORITHM_3DES = "DESede";
	public static final String ALGORITHM_BLOWFISH = "Blowfish";
	
//	private static long L200801010000 = 1199116800000l;
	private static int sequenceSeed = 0;
	private static Object sequenceLock = new Object();

	private static int sequenceSeed3 = 0;
	private static Object sequenceLock3 = new Object();
	
    private static int letterSeed = 0;
    private static Object letterLock = new Object();


	/** 获取由JDK1.5提供的32位唯一标识,并大写 */
	public static String randomUuid()
	{
	    return UUID.randomUUID().toStringValue().toUpperCase();
	}
	
	/** 获取19位yyyyMMddHHmmss + 5位循环使用的sequence的长整数，支持1秒内作100000次操作不重复 */
	public static long randomLongId()
	{
	    String datetime14 = DateTimeUtil.getDateTime14String();
	    String sequence5 = getSequence5();
	    
	    return Long.parseLong(datetime14 + sequence5);
	}

	public static long randomLong15Id()
	{
	    String datetime12 = DateTimeUtil.getDateTime12String();
	    String sequence5 = getSequence3();

	    return Long.parseLong(datetime12 + sequence5);
	}
	
	/** 获取16位yyMMddHHmmss + 4(a-z0-9)的循环字符串，支持在1秒内作4个小写字母和数字的随机 */
	public static String random16str()
	{
	    String datetime12 = DateTimeUtil.getDateTime12String();
	    String sequence4 = getLetter4();
	    
	    return datetime12 + sequence4;
	}
	
    private static String getSequence5()
    {//从0开始
        synchronized (sequenceLock)
        {
            if (sequenceSeed > 99999)
                sequenceSeed = 0;
            
            int value = sequenceSeed++;
            return StringUtil.getPrefixFixLenStr(value, 5, '0');
        }
    }

    private static String getSequence3()
    {//从0开始
        synchronized (sequenceLock3)
        {
            if (sequenceSeed3 > 99999)
                sequenceSeed3 = 0;

            int value = sequenceSeed3++;
            return StringUtil.getPrefixFixLenStr(value, 3, '0');
        }
    }
    
    private static String getLetter4()
    {//从1开始
        synchronized (letterLock)
        {
            if (letterSeed > 1679616)
                letterSeed = 0;
            
            int value = letterSeed++;
            String str = "";
            for (int i=0;i<4;i++)
            {
                int divisor = value / 36;
                int remainder = value % 36;
                char c = StringUtil.LETTERS_DIGITS_LOWERCASE.charAt(remainder);
                str = c + str;
                value = divisor;
            }
            
            return str;
        }
    }


    
    /**
     * URL UTF-8编码
     * 
     * @param src 原字符串
     * @return 结果字符串
     */
    public static String urlEncodeUTF8(String src)
    {
        try
        {
            return URLEncoder.encode(src, "UTF-8");
        }
        catch (UnsupportedEncodingException e)
        {
            return null;
        }
    }
    
    /**
     * MD5 编码,返回本地编码字符串
     * 
     * @param src 源串
     * @return 目标串
     */
     public static String md5Encoder(String src, String encoding) 
     {
         if (src == null)
             return null;
         
         byte[] destBytes = md5EncoderByte(src, encoding);
         String dest = "";
         for (int i = 0; i < 16; i++) 
         {
             dest += byteToHEX(destBytes[i]);
         }
         
         return dest;
     }


     
	/**
	 * MD5 编码,返回本地编码字符串
	 * 
	 * @param src 源串
	 * @return 目标串
	 */
	 public static String md5Encoder(String src) 
	 {
		 return md5Encoder(src, null);
	 }

	/**
	 * MD5 编码,返回本地编码字符串
	 *
	 * @param src 源串
	 * @return 目标串
	 */
	 public static String md5Salt(String src)
	 {
		 return md5Encoder(String.format("%s_%s", src,MD5_SALT), null);
	 }
	 
	   /**
      * MD5编码,返回byte数组,本地编码
      * 
      * @param src 源串
      * @return 目标编码
      */
     public static byte[] md5EncoderByte(String src)
     {
         return md5EncoderByte(src, null);
     }
     
	 /**
	  * MD5编码,返回byte数组
	  * 
	  * @param src 源串
	  * @return 目标编码
	  */
	 public static byte[] md5EncoderByte(String src, String encoding)
	 {
		 if (src == null)
			 return null;
		 
        try
        {
            if (encoding == null)
                return md5EncoderByte(src.getBytes());
            else
                return md5EncoderByte(src.getBytes(encoding));
        }
        catch (UnsupportedEncodingException e)
        {
            e.printStackTrace();
            return null;
        }
	 }

	 /**
	  * MD5编码,返回byte数组
	  * 
	  * @param buf 源byte数组
	  * @return 目标编码
	  */
	 public static byte[] md5EncoderByte(byte[] buf)
	 {
		 if (buf == null)
			 return null;
		 
		 try
		 {
			 MessageDigest md5Temp = MessageDigest.getInstance(ALGORITHM_MD5);
             return md5Temp.digest(buf);
		 }
		 catch(Exception e)
		 {
			 e.printStackTrace();
			 return null;
		 }
	}
	 
	/**
	 * SHA 编码,返回本地编码字符串
	 * 
	 * @param src 源串
	 * @return 目标串
	 */
	 public static String shaEncoder(String src) 
	 {
		 if (src == null)
			 return null;
		 
         byte[] destBytes = shaEncoderByte(src);
         String dest = "";
         for (int i = 0; i < 16; i++) 
         {
             dest += byteToHEX(destBytes[i]);
         }
         
         return dest;
	}
	
	 /**
	  * SHA编码,返回byte数组
	  * 
	  * @param src 源串
	  * @return 目标编码
	  */
	 public static byte[] shaEncoderByte(String src)
	 {
		 if (src == null)
			 return null;
		 
        try
        {
            return shaEncoderByte(src.getBytes("GBK"));
        }
        catch (UnsupportedEncodingException e)
        {
            e.printStackTrace();
            return null;
        }
	 }

	 /**
	  * SHA编码,返回byte数组
	  * 
	  * @param buf 源byte数组
	  * @return 目标编码
	  */
	 public static byte[] shaEncoderByte(byte[] buf)
	 {
         if (buf == null)
             return null;
         
         try
         {
             MessageDigest shaTemp = MessageDigest.getInstance(ALGORITHM_SHA);
             return shaTemp.digest(buf);
         }
         catch(Exception e)
         {
             e.printStackTrace();
             return null;
         }
	}
	 
	/**
	 * Base64 编码
	 * 
	 * @param src 源串
	 * @return 目标串
	 */
	public static String base64Encode(String src) 
	{
		if (src == null)
			return null;
		
		return base64Encode(src.getBytes());
	}
    
    /**
     * Base64 编码
     * 
     * @param src 源串
     * @return 目标串
     */
    public static String base64Encode(byte[] src) 
    {
        if (src == null)
            return null;
        
        return new String(Base64.encodeBase64(src));
    }

	   /**
     * Base64 编码
     * 
     * @param src 源串
     * @return 目标串
     */
    public static String base64Encode(String src, String encoding) 
    {
        if (src == null)
            return null;
        
        try
        {
            byte[] dest = Base64.encodeBase64(src.getBytes(encoding));
            
            return new String(dest, encoding);
        }
        catch (UnsupportedEncodingException e)
        {
            throw new IllegalArgumentException(e);
        }
    }

    
    /**
     * Base64 解码
     * 
     * @param dest 目标串
     * @return 源编码
     */
    public static byte[] base64DecodeBytes(String dest) 
    {
        if (dest == null)
            return null;
        
        return Base64.decodeBase64(dest.getBytes());
    }
    
	/**
	 * Base64 解码
	 * 
	 * @param dest 目标串
	 * @return 源串
	 */
	public static String base64Decode(String dest) 
	{
		if (dest == null)
			return null;
		
		byte[] b = Base64.decodeBase64(dest.getBytes());
		return new String(b);
	}

    /**
     * Base64 解码
     * 
     * @param dest 目标串
     * @return 源串
     */
    public static String base64Decode(String dest, String encoding) 
    {
        if (dest == null)
            return null;
        try
        {
            byte[] b = Base64.decodeBase64(dest.getBytes(encoding));
            return new String(b, encoding);
        }
        catch (UnsupportedEncodingException e)
        {
            throw new IllegalArgumentException(e);
        }
    }

	/**
	 * Base64 解码
	 * 
	 * @param dest 目标串
	 * @return 源编码
	 */
	public static byte[] base64DecodeBytes(String dest, String encoding) 
	{
		if (dest == null)
			return null;
		
		try
		{
		    return Base64.decodeBase64(dest.getBytes(encoding));
        }
        catch (UnsupportedEncodingException e)
        {
            throw new IllegalArgumentException(e);
        }
	}
	

	/**
	 * DES编码
	 * 
	 * @param key 密钥
	 * @param src 源
	 * @return 目标码
	 */
	public static byte[] desEncrypt(byte[] key, byte[] src)
	{
		try 
		{
			DES des = new DES(key);
			return des.encrypt(src);
		} 
		catch (Exception e) 
		{
			e.printStackTrace();
			return null;
		}
	}

	
	/**
	 * 3DES编码
	 * 
	 * @param key
	 * @param src
	 * @return 目标码
	 */
	public static byte[] tripleDesEncrypt(byte[] key, byte[] src)
	{
		try 
		{
		    byte[] key1 = NetBits.getBytes(key, 0, 8);
	        byte[] key2 = NetBits.getBytes(key, 8, 8);
	        byte[] key3 = null;
	        if (key.length == 16)
	            key3 = NetBits.getBytes(key, 0, 8);
	        else
	            key3 = NetBits.getBytes(key, 16, 8);
	        
	        DES des1 = new DES(key1);
	        byte[] dest = des1.encrypt(src);
	        
	        DES des2 = new DES(key2);
	        dest = des2.decrypt(dest);
	        
	        DES des3 = new DES(key3);
	        dest = des3.encrypt(dest);
	        
	        return dest;
		} 
		catch (Exception e) 
		{
			e.printStackTrace();
			return null;
		}
		

	}
	
	/**
	 * DES解码
	 * 
	 * @param key 密钥
	 * @param src 源
	 * @return 目标码
	 */
	public static byte[] desDecrypt(byte[] key, byte[] src)
	{
        try 
        {
            DES des = new DES(key);
            return des.decrypt(src);
        } 
        catch (Exception e) 
        {
            e.printStackTrace();
            return null;
        }
	}

	public static byte[] tripleDesDecrypt(byte[] key, byte[] src)
	{
	    try 
        {
	        byte[] key1 = NetBits.getBytes(key, 0, 8);
	        byte[] key2 = NetBits.getBytes(key, 8, 8);
	        byte[] key3 = null;
            if (key.length == 16)
                key3 = NetBits.getBytes(key, 0, 8);
            else
                key3 = NetBits.getBytes(key, 16, 8);
            
	        DES des3 = new DES(key1);
	        byte[] dest = des3.decrypt(src);
	        
	        DES des2 = new DES(key2);
	        dest = des2.encrypt(dest);
	        
	        DES des1 = new DES(key3);
	        dest = des1.decrypt(dest);
	        
	        return dest;
        } 
        catch (Exception e) 
        {
            e.printStackTrace();
            return null;
        }
	}
	
    /**
     * 把字节转换为16进制字符
     * 
     * @param ib
     * @return
     */
	public static String byteToHEX(byte ib) 
	{
	    char[] Digit = {'0','1','2','3','4','5','6','7','8','9','A','B','C','D','E','F'};
	    char [] ob = new char[2];
	    ob[0] = Digit[(ib >>> 4) & 0X0F];
	    ob[1] = Digit[ib & 0x0F];
	    String s = new String(ob);
	    return s;
    }
	/**
	 * s随机生成一个6位数的账号
	 */
	public static String randomName(){
		String val = "";
		Random random = new Random();
		//参数length，表示生成几位随机数
		for(int i = 0; i < 6; i++) {
			String charOrNum = random.nextInt(2) % 2 == 0 ? "char" : "num";
			//输出字母还是数字
			if( "char".equalsIgnoreCase(charOrNum) ) {
				//输出小写字母
				val += (char)(random.nextInt(26) + 97);
			} else if( "num".equalsIgnoreCase(charOrNum) ) {
				//输出数字
				val += String.valueOf(random.nextInt(10));
			}
		}
		return val;
	}

	/**
	 * 达到 四位随机 验证码
	 *
	 * @return
	 */
    public static String getCaptchaRandomNum(){
		String str="ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789";
		StringBuilder sb=new StringBuilder(4);
		for(int i=0;i<4;i++)
		{
			char ch=str.charAt(new Random().nextInt(str.length()));
			sb.append(ch);
		}
		return sb.toString();
	}
    public static void main(String[] args)
    {

    }
    
}

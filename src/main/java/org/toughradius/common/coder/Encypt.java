package org.toughradius.common.coder;

/**
 * 加密算法,主要用于数据库密码的加密
 */
public class Encypt
{
    private static String KEY = "!@#$%^";
    private static int KEY_LEN = 6;

    /**
     * 加密,给定一个字符串，通过作^转成16进制方式加密
     * 
     * @param str 加密前字符串
     * @return 加密后字符串
     */
    public static String encrypt(String str)
    {
        StringBuffer strb = new StringBuffer();
        for(int i=0;i<str.length();i++)
        {
            char key = KEY.charAt(i % KEY_LEN);
            int ch = str.charAt(i)^key;
            String hex = Integer.toHexString(ch).toUpperCase();
            if (hex.length() == 1)
                hex = "0" + hex;
            strb.append(hex);
        }   
        
        return strb.toString();
    }
    
    /**
     * 解密,给定一个字符串，通过加密的逆操作进行解密
     * 
     * @param str 解密前字符串
     * @return 解密后字符串
     */
    public static String decrypt(String str)
    {
        StringBuffer strb = new StringBuffer();

        if(str.length() %2 != 0) 
            str ="0" + str;   
        
        for(int i=0;i<str.length();i+=2)
        {
            char key = KEY.charAt((i/2) % KEY_LEN);
            int b = hex2byte(str.charAt(i + 1));     
            b = (b + 0x10 * hex2byte(str.charAt(i)));
            b = b ^ key;
            strb.append((char)b);
        }
        return strb.toString();
    }
    
    /** 十六进行字符转字节 */
    private static int hex2byte(char c)
    {
        if (Character.isDigit(c))
            return c - '0';
        else if (Character.isLowerCase(c))
            return 10 + c - 'a';
        else
            return 10 + c - 'A';
    }
}

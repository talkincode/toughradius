package org.toughradius.common.bits;

import java.io.EOFException;
import java.io.IOException;
import java.io.InputStream;
import java.io.UnsupportedEncodingException;


/**
 * Bit组装类
 */
public class Bits
{
	public String encoding = "UTF-8";
	
	public Bits()
	{
	}
	
	public Bits(String encoding)
	{
	    this.encoding = encoding;
	}
	
	/** 设置编码 */
	public void setEncoding(String encoding)
	{
	    this.encoding = encoding;
	}
	
    /** 获取编码后的字符串长度 */
    public int getByteLen(String str)
    {
        if (str == null)
            return 0;
        
        try
        {
            return str.getBytes(encoding).length;
        }
        catch(Exception e)
        {
            return str.getBytes().length;
        }
    }
    
    /**********************************************************/
    //以下方法为在给定的偏移位,从指定的字节数组获取数据的方法
    /**********************************************************/
    
    /**
     * 从字节数组b中 根据偏移量offset读起一个byte<br>
     * 
     * @param b 字节数组
     * @param offset 偏移量
     * @return byte 返回一个byte
     */
    public byte getByte(byte[] b, OffSet offset)
    {        
        int off = offset.getOff();     
        offset.setOff(off + 1);
        
        return b[off];
    }
    
    /**
     * 从字节数组b中 根据偏移量offset读起一个boolean<br>
     * 
     * @param b 字节数组
     * @param offset 偏移量
     * @return byte 返回一个boolean
     */
    public boolean getBoolean(byte[] b, OffSet offset)
    {
        int off = offset.getOff();
        offset.setOff(off + 1);
        
        return b[off] != 0;
    }

    /**
     * 从字节数组b中 根据偏移量offset读起一个 <b>单字节</b> 的char<br>
     * 
     * @param b 字节数组
     * @param offset 偏移量
     * @return byte 返回一个char
     */
    public char getChar1(byte[] b, OffSet offset)
    {
        int off = offset.getOff();
        offset.setOff(off + 1);
        
        return (char) ((b[off + 0] & 0xFF));
    }

    /**
     * 从字节数组b中 根据偏移量offset读起一个 <b>双字节</b> 的char<br>
     * 
     * @param b 字节数组
     * @param offset 偏移量
     * @return byte 返回一个char
     */
    public char getChar2(byte[] b, OffSet offset)
    {
        int off = offset.getOff();
        offset.setOff(off + 2);
        
        return (char) (((b[off + 1] & 0xFF) << 0) + ((b[off + 0] & 0xFF) << 8));
    }
    
    /**
     * 从字节数组b中 根据偏移量offset读起一个short<br>
     * 
     * @param b 字节数组
     * @param offset 偏移量
     * @return byte 返回一个short
     */
    public short getShort(byte[] b, OffSet offset)
    {
        int off = offset.getOff();
        offset.setOff(off + 2);
        
        return (short) (((b[off + 1] & 0xFF) << 0) + ((b[off + 0] & 0xFF) << 8));
    }

    /**
     * 从字节数组b中 根据偏移量offset读起一个int<br>
     * 
     * @param b 字节数组
     * @param offset 偏移量
     * @return byte 返回一个int
     */
    public int getInt(byte[] b, OffSet offset)
    {
        int off = offset.getOff();
        offset.setOff(off + 4);
        
        return ((b[off + 3] & 0xFF) << 0) + ((b[off + 2] & 0xFF) << 8)
            + ((b[off + 1] & 0xFF) << 16) + ((b[off + 0] & 0xFF) << 24);
    }

    /**
     * 从字节数组b中 根据偏移量offset读起一个float<br>
     * 
     * @param b 字节数组
     * @param offset 偏移量
     * @return byte 返回一个float
     */
    public float getFloat(byte[] b, OffSet offset)
    {
        int off = offset.getOff();
        offset.setOff(off + 4);
        
        int i = ((b[off + 3] & 0xFF) << 0) + ((b[off + 2] & 0xFF) << 8)
            + ((b[off + 1] & 0xFF) << 16) + ((b[off + 0] & 0xFF) << 24);
        return Float.intBitsToFloat(i);
    }

    /**
     * 从字节数组b中 根据偏移量offset读起一个long<br>
     * 
     * @param b 字节数组
     * @param offset 偏移量
     * @return byte 返回一个long
     */
    public long getLong(byte[] b, OffSet offset)
    {
        int off = offset.getOff();
        offset.setOff(off + 8);
        
        return ((b[off + 7] & 0xFFL) << 0) + ((b[off + 6] & 0xFFL) << 8)
            + ((b[off + 5] & 0xFFL) << 16) + ((b[off + 4] & 0xFFL) << 24)
            + ((b[off + 3] & 0xFFL) << 32) + ((b[off + 2] & 0xFFL) << 40)
            + ((b[off + 1] & 0xFFL) << 48) + ((b[off + 0] & 0xFFL) << 56);
    }

    /**
     * 从字节数组b中 根据偏移量offset读起一个double<br>
     * 
     * @param b 字节数组
     * @param offset 偏移量
     * @return byte 返回一个double
     */
    public double getDouble(byte[] b, OffSet offset)
    {
        int off = offset.getOff();
        offset.setOff(off + 8);
        
        long j = ((b[off + 7] & 0xFFL) << 0) + ((b[off + 6] & 0xFFL) << 8)
            + ((b[off + 5] & 0xFFL) << 16) + ((b[off + 4] & 0xFFL) << 24)
            + ((b[off + 3] & 0xFFL) << 32) + ((b[off + 2] & 0xFFL) << 40)
            + ((b[off + 1] & 0xFFL) << 48) + ((b[off + 0] & 0xFFL) << 56);
        return Double.longBitsToDouble(j);
    }

    /**
     * 从字节数组b中 根据偏移量offset读起一个定长为len的byte[]<br>
     * 
     * @param b 字节数组
     * @param offset 偏移量
     * @param len 读起的长度
     * @return byte 返回一个byte[]
     */
    public byte[] getBytes(byte[] b, OffSet offset, int len)
    {        
        int off = offset.getOff();     

        byte[] bytes = new byte[len];
        if (len + off <= b.length)
            System.arraycopy(b, off, bytes, 0, len);
        else
            System.arraycopy(b, off, bytes, 0, b.length-off);
        
        offset.setOff(off + len);
        return bytes;
    }

    /**
     * 从字节数组b中 根据偏移量offset读起一个到b末端的byte[]<br>
     * 
     * @param b 字节数组
     * @param offset 偏移量
     * @return byte 返回一个byte[]
     */
    public byte[] getBytes(byte[] b, OffSet offset)
    {
        int off = offset.getOff();     
        byte[] bytes = new byte[b.length-off];
        System.arraycopy(b, off, bytes, 0, b.length-off);
        
        offset.setOff(off + bytes.length);
        return bytes;
    }
    
    /**
     * 从字节数组b中 根据偏移量offset读起一个到结束符为value的String<br>
     * 
     * @param b 字节数组
     * @param offset 偏移量
     * @param value 结束符
     * @return byte 返回一个String
     */
    public String getString(byte[] b, OffSet offset, byte value)
    {
        int off = offset.getOff();
        int i = off;
        int allLen = b.length;
        while (i < allLen)
        {
            if (b[i] == value)
                break;
            
            i++;
        }
        
        byte[] tmp = new byte[i-off];
        System.arraycopy(b, off, tmp, 0, i-off);
        String str = null;
		try
		{
			str = new String(tmp, encoding);
		}
		catch (UnsupportedEncodingException e)
		{
			str = new String(tmp);
		}

        offset.setOff(i + 1);
        return str;
    }
    
    /**
     * 从字节数组b中 根据偏移量offset读起一个定长的String<br>
     * 
     * @param b 字节数组
     * @param offset 偏移量
     * @param len 长度
     * @return byte 返回一个String
     */
    public String getString(byte[] b, OffSet offset, int len)
    {
        int off = offset.getOff();      
        int i = 0;
        while (i < len)
        {
            if (off + i >= b.length)
                break;
            if (0 == b[off+i])
                break;
            i++;
        }
        
        byte[] tmp = new byte[i];
        System.arraycopy(b, off, tmp, 0, i);
        String str = null;
		try
		{
			str = new String(tmp, encoding);
		}
		catch (UnsupportedEncodingException e)
		{
			str = new String(tmp);
		}

        offset.setOff(off + len);
        return str;
    }
    
    /**
     * 从字节数组b中 根据偏移量offset读起一个定长的String<br>
     * 
     * @param b 字节数组
     * @param offset 偏移量
     * @param len 长度
     * @param hasEnd 有结束符
     * @return byte 返回一个String
     */
    public String getString(byte[] b, OffSet offset, int len, boolean hasEnd)
    {
        int off = offset.getOff();      
        int i = 0;
        while (i < len)
        {
            if (off + i >= b.length)
                break;
            if (0 == b[off+i])
                break;
            i++;
        }
        
        byte[] tmp = new byte[i];
        System.arraycopy(b, off, tmp, 0, i);
        String str = null;
		try
		{
			str = new String(tmp, encoding);
		}
		catch (UnsupportedEncodingException e)
		{
			str = new String(tmp);
		}

        if (hasEnd)
        	offset.setOff(off + len + 1);
        else
            offset.setOff(off + len);
        
        return str;
    }
    /**********************************************************/
    //以下方法为在给定的偏移位,给定的值,插入到指定的字节数组的方法
    /**********************************************************/
    
    /**
     * 插入一个字节<br>
     * 
     * @param b 字节数组
     * @param offset 偏移量
     * @param val
     */
    public void putByte(byte[] b, OffSet offset, byte val)
    {
        int off = offset.getOff();
        b[off] = val;
        
        offset.setOff(off + 1);
    }
    
    /**
     * 插入一个boolean<br>
     * 
     * @param b 字节数组
     * @param offset
     * @param val
     */
    public void putBoolean(byte[] b, OffSet offset, boolean val)
    {
        int off = offset.getOff();
        b[off] = (byte) (val ? 1 : 0);
        
        offset.setOff(off + 1);
    }

    /**
     * 插入一个 <b>单字节</b> 的char<br>
     * 
     * @param b 字节数组
     * @param offset 偏移量
     * @param val 单字节char
     */
    public void putChar1(byte[] b, OffSet offset, char val)
    {
        int off = offset.getOff();
        b[off + 0] = (byte) (val & 0xFF);
        
        offset.setOff(off + 1);
    }
    
    /**
     * 插入一个 <b>双字节</b> 的char<br>
     * 
     * @param b 字节数组
     * @param offset 偏移量
     * @param val char 双字节char
     */
    public void putChar2(byte[] b, OffSet offset, char val)
    {
        int off = offset.getOff();
        b[off + 1] = (byte) (val >>> 0);
        b[off + 0] = (byte) (val >>> 8);
        
        offset.setOff(off + 2);
    }
    
    /**
     * 插入一个short<br>
     * 
     * @param b 字节数组
     * @param offset 偏移量
     * @param val short
     */
    public void putShort(byte[] b, OffSet offset, short val)
    {
        int off = offset.getOff();
        b[off + 1] = (byte) (val >>> 0);
        b[off + 0] = (byte) (val >>> 8);
        
        offset.setOff(off + 2);
    }

    /**
     * 插入一个int<br>
     * 
     * @param b 字节数组
     * @param offset 偏移量
     * @param val int
     */
    public void putInt(byte[] b, OffSet offset, int val)
    {
        int off = offset.getOff();
        b[off + 3] = (byte) (val >>> 0);
        b[off + 2] = (byte) (val >>> 8);
        b[off + 1] = (byte) (val >>> 16);
        b[off + 0] = (byte) (val >>> 24);
        
        offset.setOff(off + 4);
    }

    /**
     * 插入一个float<br>
     * 
     * @param b 字节数组
     * @param offset 偏移量
     * @param val float
     */
    public void putFloat(byte[] b, OffSet offset, float val)
    {
        int off = offset.getOff();
        int i = Float.floatToIntBits(val);
        b[off + 3] = (byte) (i >>> 0);
        b[off + 2] = (byte) (i >>> 8);
        b[off + 1] = (byte) (i >>> 16);
        b[off + 0] = (byte) (i >>> 24);
        
        offset.setOff(off + 4);
    }

    /**
     * 插入一个long<br>
     * 
     * @param b 字节数组
     * @param offset 偏移量
     * @param val long
     */
    public void putLong(byte[] b, OffSet offset, long val)
    {
        int off = offset.getOff();
        b[off + 7] = (byte) (val >>> 0);
        b[off + 6] = (byte) (val >>> 8);
        b[off + 5] = (byte) (val >>> 16);
        b[off + 4] = (byte) (val >>> 24);
        b[off + 3] = (byte) (val >>> 32);
        b[off + 2] = (byte) (val >>> 40);
        b[off + 1] = (byte) (val >>> 48);
        b[off + 0] = (byte) (val >>> 56);
        
        offset.setOff(off + 8);
    }

    /**
     * 插入一个double<br>
     * 
     * @param b 字节数组
     * @param offset 偏移量
     * @param val double
     */
    public void putDouble(byte[] b, OffSet offset, double val)
    {
        int off = offset.getOff();
        long j = Double.doubleToLongBits(val);
        b[off + 7] = (byte) (j >>> 0);
        b[off + 6] = (byte) (j >>> 8);
        b[off + 5] = (byte) (j >>> 16);
        b[off + 4] = (byte) (j >>> 24);
        b[off + 3] = (byte) (j >>> 32);
        b[off + 2] = (byte) (j >>> 40);
        b[off + 1] = (byte) (j >>> 48);
        b[off + 0] = (byte) (j >>> 56);
        
        offset.setOff(off + 8);
    }
    
    /**
     * 插入一个不定长字节数组<br>
     * 
     * @param b 字节数组
     * @param offset 偏移量
     * @param src 要插入的字符数组
     */
    public void putBytes(byte[] b, OffSet offset, byte[] src)
    {
        int off = offset.getOff();
        System.arraycopy(src, 0, b, off,src.length);
        
        offset.setOff(off + src.length);
    }
    
    /**
     * 插入一个定长字节数组<br>
     * 
     * @param b 字节数组
     * @param offset 偏移量
     * @param src 要插入的字符数组
     * @param len 长度
     */
    public void putBytes(byte[] b, OffSet offset, byte[] src, int len)
    {
        int off = offset.getOff();
        System.arraycopy(src, 0, b, off,len);
        
        offset.setOff(off + len);
    }
    
    /**
     * 插入一个不定长的String,由str.getBytes().length来计算,无结束符<br>
     * 
     * @param b 字节数组
     * @param offset 偏移量
     * @param str 字节串
     */
    public void putString(byte[] b, OffSet offset, String str)
    {
        int off = offset.getOff();
        if (str == null)
            str = "";
        
        byte[] ret = null;
		try
		{
			ret = str.getBytes(encoding);
		}
		catch (UnsupportedEncodingException e)
		{
			ret = str.getBytes();
		}            
		
        System.arraycopy(ret, 0, b, off, ret.length);

        offset.setOff(off + ret.length);
    }

    /**
     * 插入一个定长的String,由str.getBytes().length来计算,以endValue结束<br>
     * 
     * @param b 字节数组
     * @param offset 偏移量
     * @param str 字节串
     * @param endValue 结束符
     */
    public void putString(byte[] b, OffSet offset, String str, byte endValue)
    {
        int off = offset.getOff();
        if (str == null)
            str = "";
        
        byte[] ret = null;
		try
		{
			ret = str.getBytes(encoding);
		}
		catch (UnsupportedEncodingException e)
		{
			ret = str.getBytes();
		}            
		          
        System.arraycopy(ret, 0, b, off, ret.length);
        b[off + ret.length] = endValue;
          
        offset.setOff(off + ret.length + 1);
    }
    
    /**
     * 插入一个定长的String,无结束符<br>
     * 
     * @param b 字节数组
     * @param offset 偏移量
     * @param str 字节串
     * @param len 长度
     */
    public void putString(byte[] b, OffSet offset, String str, int len)
    {
        int off = offset.getOff();
        if (str == null)
            str = "";
        
        byte[] ret = null;
		try
		{
			ret = str.getBytes(encoding);
		}
		catch (UnsupportedEncodingException e)
		{
			ret = str.getBytes();
		}            
		        
        if (ret.length > len)
            System.arraycopy(ret, 0, b, off, len);
        else
            System.arraycopy(ret, 0, b, off, ret.length);

        offset.setOff(off + len);
    }
    
    /**
     * 插入一个定长的String,并以一个byte结束<br>
     * 
     * @param b 字节数组
     * @param offset 偏移量
     * @param str 字节串
     * @param len 长度
     * @param endValue 结束符
     */
    public void putString(byte[] b, OffSet offset, String str, int len, byte endValue)
    {
        int off = offset.getOff();
        if (str == null)
            str = "";
        
        byte[] ret = null;
		try
		{
			ret = str.getBytes(encoding);
		}
		catch (UnsupportedEncodingException e)
		{
			ret = str.getBytes();
		}            
		
        if (ret.length > len)
            System.arraycopy(ret, 0, b, off, len);
        else
            System.arraycopy(ret, 0, b, off, ret.length);

        b[off + ret.length] = endValue;
        offset.setOff(off + len + 1);
    }
    
    /******************************************/
    //以下方法为在给定的int型偏移位,给定的值,插入到指定的字节数组的方法
    /******************************************/
    
    /*
     * Methods for unpacking primitive values from byte arrays starting at given
     * offsets.
     */

    public boolean getBoolean(byte[] b, int off)
    {
        return b[off] != 0;
    }

    public char getChar1(byte[] b, int off)
    {
        return (char) ((b[off + 0] & 0xFF));
    }

    public char getChar2(byte[] b, int off)
    {
        return (char) (((b[off + 1] & 0xFF) << 0) + ((b[off + 0] & 0xFF) << 8));
    }
    
    public short getShort(byte[] b, int off)
    {
        return (short) (((b[off + 1] & 0xFF) << 0) + ((b[off + 0] & 0xFF) << 8));
    }

    public int getInt(byte[] b, int off)
    {
        return ((b[off + 3] & 0xFF) << 0) + ((b[off + 2] & 0xFF) << 8)
            + ((b[off + 1] & 0xFF) << 16) + ((b[off + 0] & 0xFF) << 24);
    }

    public float getFloat(byte[] b, int off)
    {
        int i = ((b[off + 3] & 0xFF) << 0) + ((b[off + 2] & 0xFF) << 8)
            + ((b[off + 1] & 0xFF) << 16) + ((b[off + 0] & 0xFF) << 24);
        return Float.intBitsToFloat(i);
    }

    public long getLong(byte[] b, int off)
    {
        return ((b[off + 7] & 0xFFL) << 0) + ((b[off + 6] & 0xFFL) << 8)
            + ((b[off + 5] & 0xFFL) << 16) + ((b[off + 4] & 0xFFL) << 24)
            + ((b[off + 3] & 0xFFL) << 32) + ((b[off + 2] & 0xFFL) << 40)
            + ((b[off + 1] & 0xFFL) << 48) + ((b[off + 0] & 0xFFL) << 56);
    }

    public double getDouble(byte[] b, int off)
    {
        long j = ((b[off + 7] & 0xFFL) << 0) + ((b[off + 6] & 0xFFL) << 8)
            + ((b[off + 5] & 0xFFL) << 16) + ((b[off + 4] & 0xFFL) << 24)
            + ((b[off + 3] & 0xFFL) << 32) + ((b[off + 2] & 0xFFL) << 40)
            + ((b[off + 1] & 0xFFL) << 48) + ((b[off + 0] & 0xFFL) << 56);
        return Double.longBitsToDouble(j);
    }

    public String getString(byte[] b, int off, byte value)
    {
        int i = off;
        int allLen = b.length;
        while (i < allLen)
        {
            if (b[i] == value)
                break;
            
            i++;
        }
        
        byte[] tmp = new byte[i-off];
        System.arraycopy(b, off, tmp, 0, i-off);
        
        String str = null;
        try
        {
            str = new String(tmp, encoding);
        }
        catch (UnsupportedEncodingException e)
        {
            str = new String();
        }

        return str;
    }
    
    public String getString(byte[] b, int off, int len)
    {
        int i = 0;
        while (i < len)
        {
            if (off + i >= b.length)
                break;
            if (0 == b[off+i])
                break;
            i++;
        }
        
        byte[] tmp = new byte[i];
        System.arraycopy(b, off, tmp, 0, i);
        
        String str = null;
		try
		{
			str = new String(tmp, encoding);
		}
		catch (UnsupportedEncodingException e)
		{
			str = new String();
		}

        return str;
    }
    
    public byte getByte(byte[] b, int off)
    {
        return b[off];
    }
    
    public byte[] getBytes(byte[] b, int off, int len)
    {
        byte[] bytes = new byte[len];
        if (len + off <= b.length)
            System.arraycopy(b, off, bytes, 0, len);
        else
            System.arraycopy(b, off, bytes, 0, b.length-off);
        
        return bytes;
    }

    public byte[] getBytes(byte[] b, int off)
    {
        byte[] bytes = new byte[b.length-off];
        System.arraycopy(b, off, bytes, 0, b.length-off);
        
        return bytes;
    }
    
    /*
     * Methods for packing primitive values into byte arrays starting at given
     * offsets.
     */

    public int putBoolean(byte[] b, int off, boolean val)
    {
        b[off] = (byte) (val ? 1 : 0);
        
        return off + 1;
    }

    public int putChar1(byte[] b, int off, char val)
    {
        b[off + 0] = (byte) (val & 0xFF);
        
        return off + 1;
    }
    
    public int putChar2(byte[] b, int off, char val)
    {
        b[off + 1] = (byte) (val >>> 0);
        b[off + 0] = (byte) (val >>> 8);
        
        return off + 2;
    }
    
    public int putShort(byte[] b, int off, short val)
    {
        b[off + 1] = (byte) (val >>> 0);
        b[off + 0] = (byte) (val >>> 8);
        
        return off + 2;
    }

    public int putInt(byte[] b, int off, int val)
    {
        b[off + 3] = (byte) (val >>> 0);
        b[off + 2] = (byte) (val >>> 8);
        b[off + 1] = (byte) (val >>> 16);
        b[off + 0] = (byte) (val >>> 24);
        
        return off + 4;
    }

    public int putFloat(byte[] b, int off, float val)
    {
        int i = Float.floatToIntBits(val);
        b[off + 3] = (byte) (i >>> 0);
        b[off + 2] = (byte) (i >>> 8);
        b[off + 1] = (byte) (i >>> 16);
        b[off + 0] = (byte) (i >>> 24);
        
        return off + 4;
    }

    public int putLong(byte[] b, int off, long val)
    {
        b[off + 7] = (byte) (val >>> 0);
        b[off + 6] = (byte) (val >>> 8);
        b[off + 5] = (byte) (val >>> 16);
        b[off + 4] = (byte) (val >>> 24);
        b[off + 3] = (byte) (val >>> 32);
        b[off + 2] = (byte) (val >>> 40);
        b[off + 1] = (byte) (val >>> 48);
        b[off + 0] = (byte) (val >>> 56);
        
        return off + 8;
    }

    public int putDouble(byte[] b, int off, double val)
    {
        long j = Double.doubleToLongBits(val);
        b[off + 7] = (byte) (j >>> 0);
        b[off + 6] = (byte) (j >>> 8);
        b[off + 5] = (byte) (j >>> 16);
        b[off + 4] = (byte) (j >>> 24);
        b[off + 3] = (byte) (j >>> 32);
        b[off + 2] = (byte) (j >>> 40);
        b[off + 1] = (byte) (j >>> 48);
        b[off + 0] = (byte) (j >>> 56);
        
        return off + 8;
    }
    
    public int putByte(byte[] b, int off, byte val)
    {
        b[off] = val;
        
        return off + 1;
    }
    
    public int putBytes(byte[] b, int off, byte[] src)
    {
        System.arraycopy(src, 0, b, off,src.length);
        
        return off + src.length;
    }
    
    public int putBytes(byte[] b, int off, byte[] src, int len)
    {
        System.arraycopy(src, 0, b, off,len);
        
        return off + len;
    }
    
    public int putString(byte[] b, int off, String str)
    {
        if (str == null)
            str = "";
        
        byte[] ret = null;
        try
        {
            ret = str.getBytes(encoding); 
        }
        catch (UnsupportedEncodingException e)
        {
            ret = str.getBytes();
        }    
        
        System.arraycopy(ret, 0, b, off, ret.length);

        return off + ret.length;
    }

    public int putString(byte[] b, int off, String str,int len)
    {
        if (str == null)
            str = "";
        
        byte[] ret = null;
		try
		{
			ret = str.getBytes(encoding);
		}
		catch (UnsupportedEncodingException e)
		{
			ret = str.getBytes();
		}            
		          
        if (ret.length > len)
            System.arraycopy(ret, 0, b, off, len);
        else
            System.arraycopy(ret, 0, b, off, ret.length);

        return off + len;
    }
    
    /**********************************************************/
    //以下方法为String,int等与byte[]之间的转换
    /**********************************************************/
    
    public void fillBytes(byte[] bytes, int off, byte b, int len)
    {
        for (int i=0;i<bytes.length;i++)
        {
            if (i < off)
                continue;
            
            if (i == (off +len))
                break;
            bytes[i] = b;
        }
    }
    
    public String toString(byte[] bytes)
    {
        if (bytes == null)
            return null;
        
        try
        {
            return new String(bytes, encoding);
        }
        catch (UnsupportedEncodingException e)
        {
            return new String(bytes);
        }
    }
    
    public byte[] toBytes(String str)
    {
        if (str == null)
            return new byte[0];
        
        try
        {
            return str.getBytes(encoding);
        }
        catch (UnsupportedEncodingException e)
        {
            throw new RuntimeException("encoding error");
        }
    }
    
    public byte[] toBytes(String str, int len)
    {
        if (str == null)
            return new byte[0];
        
        byte[] buf = null;
        
        try
        {
            buf = str.getBytes(encoding);
        }
        catch (UnsupportedEncodingException e)
        {
            throw new RuntimeException("encoding error");
        }
        
        byte[] bytes = new byte[len];
        if (buf.length > len)
            System.arraycopy(buf,0,bytes,0,len);
        else
            System.arraycopy(buf,0,bytes,0,buf.length);
        
        return bytes;
    }
    
    public byte[] toBytes(String str, String aEncoding)
    {
        if (str == null)
            return new byte[0];
        
		try
		{
			return str.getBytes(aEncoding);
		}
		catch (UnsupportedEncodingException e)
		{
			throw new RuntimeException("encoding error");
		}    	
    }
    
    public byte[] toBytes(String str, int len, String aEncoding)
    {
        if (str == null)
            return new byte[0];
        
		try
		{
	    	byte[] buf = str.getBytes(aEncoding);
	    	byte[] bytes = new byte[len];
	    	if (buf.length > len)
	    		System.arraycopy(buf,0,bytes,0,len);
	    	else
	    		System.arraycopy(buf,0,bytes,0,buf.length);
	    	
	    	return bytes;
		}
		catch (UnsupportedEncodingException e)
		{
			throw new RuntimeException("encoding error");
		}    	
    }
    

    public String toHEXString(byte[] b)
    {
        char[] Digit = { '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'A',
            'B', 'C', 'D', 'E', 'F' };
        StringBuffer s = new StringBuffer();
        char[] ob = new char[2];
        for (int i = 0; i < b.length; i++)
        {
            byte ib = b[i];
            ob[0] = Digit[(ib >>> 4) & 0X0F];
            ob[1] = Digit[ib & 0X0F];
            s.append(new String(ob) + " ");
        }
        return s.toString();
    }

    /** HEX 转bytes, 支持HEX中间加空格*/
    public byte[] toBytesByHEX(String hex)
    {
        hex = hex.replaceAll(" ", "");
        hex = hex.toUpperCase();
        byte[] bytes = new byte[hex.length() / 2];
        for (int i=0, v=0;i<hex.length();i=i+2,v++)
        {
            byte a = (byte)(hex.charAt(i) - 0x30);
            if (a > 15)
                a -= 7;
            byte b = (byte)(hex.charAt(i+1) - 0x30);
            if (b > 15)
                b -= 7;
            bytes[v] = (byte)(a * 16 + b);
        }
        
        return bytes;
    }
    
    public String toHEXStringNoSpace(byte[] b)
    {
        char[] Digit = { '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'A',
            'B', 'C', 'D', 'E', 'F' };
        StringBuffer s = new StringBuffer();
        char[] ob = new char[2];
        for (int i = b.length - 1; i >= 0; i--)
        {
            byte ib = b[i];
            ob[0] = Digit[(ib >>> 4) & 0X0F];
            ob[1] = Digit[ib & 0X0F];
            s.append(new String(ob));
        }
        return s.toString();
    }
    
    public byte[] readStream(InputStream input, int len) throws IOException
    {
        int readcount = 0, ret = 0;
        byte[] buf = new byte[len];
        while (readcount < len)
        {
            ret = input.read(buf, readcount, len - readcount);
            if (-1 == ret)
                throw new EOFException("按长度读消息时,长度不够即到达流尾端");
            readcount += ret;
        }
        return buf;
    }
}
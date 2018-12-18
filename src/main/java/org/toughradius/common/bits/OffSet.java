/*
 * 版权所有 (C) 2005-2008 WWW.ZOULAB.COM。保留所有权利。
 * 版本：
 * 修改记录：
 *		1、2006-1-23，zouchenggang创建; 
 */
package org.toughradius.common.bits;

/**
 * 创建一个OFFSET,用于Bits偏移
 */
public class OffSet
{
	private int off;
    
    public OffSet(int off)
    {
    	this.off = off;
    }
    
    public void setOff(int off)
    {
    	this.off = off;
    }
    
    public int getOff()
    {
    	return off;
    }
}

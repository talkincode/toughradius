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

import gnu.crypto.prng.IRandom;
import gnu.crypto.prng.MDGenerator;
import gnu.crypto.prng.PRNGFactory;

import java.util.Calendar;
import java.util.GregorianCalendar;
import java.util.LinkedHashMap;
import java.util.Map;

/**
 * A Random Number Generator (wrapper) for JRadius
 *
 * @author David Bird
 */
public class RadiusRandom
{
    static final Map attrib = new LinkedHashMap();
    static final IRandom rand = PRNGFactory.getInstance("MD");
    
    static
    {
        Calendar cal = GregorianCalendar.getInstance();
        byte seed[] = Long.toBinaryString(cal.getTimeInMillis()).getBytes();
        attrib.put(MDGenerator.MD_NAME, "MD5");
        attrib.put(MDGenerator.SEEED, seed);
        rand.init(attrib);
    }
    
    /**
     * Generates an array of random bytes.
     * @param length number of random bytes to generate
     * @return array of random bytes
     */
    public static byte[] getBytes(int length)
    {
        byte result[] = new byte[length];
        synchronized (rand)
        {
            for (int i = 0; i < length; i++)
            {
                try
                {
                    result[i] ^= rand.nextByte();
                }
                catch (Exception e)
                {
                    e.printStackTrace();
                }
            }
        }
        return result;
    }
    
    public static String getRandomString(int length)
    {
        return RadiusUtils.byteArrayToHexString(getBytes(length));
    }
}
package org.toughradius.common;

import java.io.IOException;
import java.io.InputStream;
import java.io.UnsupportedEncodingException;
import java.text.DecimalFormat;
import java.util.Calendar;
import java.util.Date;
import java.util.Random;

/**
 * 字符串工具类
 */
public class StringUtil
{
    private static final String RMB_NUM[] = { "零", "壹", "贰", "叁", "肆", "伍", "陆", "柒", "捌", "玖" };
    private static final String RMB_UNIT[] = { "圆", "拾", "佰", "仟", "万", "拾", "佰", "仟", "亿", "拾","佰", "仟" };
    private static final String RMB_DEC[] = { "角", "分" };
    
    /** 数字chars */
    public static final String DIGITS = "0123456789";
    
    /**  小写字母chars */
    public static final String LETTERS_LOWERCASE = "abcdefghijklmnopqrstuvwxyz";
    
    /**  小写字母chars + 数字 */
    public static final String LETTERS_DIGITS_LOWERCASE = "0123456789abcdefghijklmnopqrstuvwxyz";

    /** 大写字母chars */
    public static final String LETTERS_UPPERCASE = "ABCDEFGHIJKLMNOPQRSTUVWXYZ";

    /** 全部字母chars */
    public static final String LETTERS = LETTERS_LOWERCASE + LETTERS_UPPERCASE;

    /** 全部字母数字 */
    public static final String LETTERS_DIGITS = LETTERS + DIGITS;
    
    /** 空白的chars (包括空格,\t,\n,\r) */
    public static final String WHITE_SPACE = " \t\n\r";
  
    private static char[] LOWER_CASES = {
        '\000','\001','\002','\003','\004','\005','\006','\007',
        '\010','\011','\012','\013','\014','\015','\016','\017',
        '\020','\021','\022','\023','\024','\025','\026','\027',
        '\030','\031','\032','\033','\034','\035','\036','\037',
        '\040','\041','\042','\043','\044','\045','\046','\047',
        '\050','\051','\052','\053','\054','\055','\056','\057',
        '\060','\061','\062','\063','\064','\065','\066','\067',
        '\070','\071','\072','\073','\074','\075','\076','\077',
        '\100','\141','\142','\143','\144','\145','\146','\147',
        '\150','\151','\152','\153','\154','\155','\156','\157',
        '\160','\161','\162','\163','\164','\165','\166','\167',
        '\170','\171','\172','\133','\134','\135','\136','\137',
        '\140','\141','\142','\143','\144','\145','\146','\147',
        '\150','\151','\152','\153','\154','\155','\156','\157',
        '\160','\161','\162','\163','\164','\165','\166','\167',
        '\170','\171','\172','\173','\174','\175','\176','\177' };
    
    /** 生成UTF-8的字符串 */
    public static String newStringByUTF8(byte[] data)
    {
        return newString(data, "UTF-8");
    }
    
    /** 生成指定编码的字符串 */
    public static String newString(byte[] data, String encoding)
    {
        try
        {
            return new String(data, encoding);
        }
        catch (UnsupportedEncodingException e)
        {
            return null;
        }
    }
    
    /** 字符串转换为int,异常不处理 注:默认10进制 */
    public static int toInt(String s)
    {
    	return Integer.parseInt(s);
    }

    /** 字符串转换为int 如果异常则返回异常值 */
    public static int toInt(String s,int exception)
    {
        try
        {
            return Integer.parseInt(s);
        }
        catch(NumberFormatException e)
        {
            return exception;
        }
    }
    
    /**
     * 把字符串数组传换成整数数组 适合在web的form表单提交时得到的是字符串数组,但实际要整数数组
     * 
     * @param array
     * @return int[] 转换后的整数数组
     */
    public static int[] toIntArray(String[] array)
    {
        int[] intArray = new int[array.length];
        for (int i = 0; i < array.length; i++)
            intArray[i] = Integer.parseInt(array[i]);
        return intArray;
    }

    /**
     * 将普通字符串格式化成数据库认可的字符串格式
     * 
     * @param str 要格式化的字符串
     * @return 合法的数据库字符串
     */
    public static String toSql(String str)
    {
        return str.replaceAll("'", "''");
    }

    /**
     * 把id数组组合为字符串 适合在组织sql语句是会用上
     * 
     * @param ids id数组
     * @param delimeter 分隔符
     * @return
     */
    public static String toStringByArray(int[] ids, String separator)
    {
        if (ids == null || ids.length == 0)
            return "";
        
        StringBuffer buf = new StringBuffer();
        for (int i = 0; i < ids.length; i++)
        {
            buf.append(ids[i]);
            if (i < ids.length - 1)
                buf.append(separator);
        }
        return buf.toString();
    }
    
    /**
     * 把id数组组合为字符串 适合在组织sql语句是会用上
     * 
     * @param ids id数组
     * @param delimeter 分隔符
     * @return
     */
    public static String toStringSqlByArray(String[] ids, String separator)
    {
        if (ids == null || ids.length == 0)
            return "";
        
        StringBuffer buf = new StringBuffer();
        for (int i = 0; i < ids.length; i++)
        {
            buf.append("'").append(ids[i]).append("'");
            
            if (i < ids.length - 1)
                buf.append(separator);
        }

        return buf.toString();
    }

    
    /**
     * 将字符串格式化成 HTML 代码输出 除普通特殊字符外，还对空格、制表符和换行进行转换， 以将内容格式化输出， 适合于 HTML 中的显示输出
     * 
     * @param str 要格式化的字符串
     * @return 格式化后的字符串
     */
    public static String toHtml(String str)
    {
        if (str == null)
            return "";

        String html = new String(str);

        html = toHtmlInput(html);
        html = html.replaceAll("\r\n", "\n");
        html = html.replaceAll("\n", "<br>\n");
        html = html.replaceAll("\t", "    ");
        html = html.replaceAll("  ", " &nbsp;");

        return html;
    }

    /**
     * 将字符串格式化成 HTML 代码输出 只转换特殊字符，适合于 HTML 中的表单区域
     * 
     * @param str 要格式化的字符串
     * @return 格式化后的字符串
     */
    public static String toHtmlInput(String str)
    {
        if (str == null)
            return "";

        String html = new String(str);

        html = html.replaceAll("&", "&amp;");
        html = html.replaceAll("<", "&lt;");
        html = html.replaceAll(">", "&gt;");
        html = html.replaceAll("\"", "&quot;");

        return html;
    }

    /**
     * 把文符转换成编辑器合法的格式
     * 
     * @param source 源字符串
     * @return String
     */
    public static String toHtmlEditor(String source)
    {
        if (source == null)
            return null;

        String html = new String(source);

        html = html.replaceAll("\"", "\\\"");
        html = html.replaceAll("\r\n", "\n");
        html = html.replaceAll("\n", "\\n");
        html = html.replaceAll("\t", "    ");
        //html = html.replaceAll(" ", " &nbsp;"); //这一项暂时不加

        html = html.replaceAll("<script", "\\<script");
        html = html.replaceAll("<SCRIPT", "\\<SCRIPT");
        html = html.replaceAll("/script>", "/script\\>");
        html = html.replaceAll("/SCRIPT>", "/SCRIPT\\>");

        return html;
    }

    /**
     * 合并字符串,关于拆分字符串为数组,请直接调用String.split(String regex)方法
     * 
     * 举例:如字符串数组有三个元素hello、my、friend，用户定义分隔符为|，那么字符串可 以合并为hello|my|friend
     * @param array 字符串数组
     * @return 合并后的字符串
     */
    public static String toStringByArray(String[] array, String separator)
    {
        if (array == null || array.length == 0)
            return null;
        
        if (ValidateUtil.isEmpty(separator))
            separator = ",";

        StringBuffer buf = new StringBuffer();
        for (int i = 0; i < array.length; i++)
        {
            buf.append(array[i]);
            if (i < array.length - 1)
                buf.append(separator);
        }

        return buf.toString();
    }

    /**
     * 替换一个序列
     * 
     * @param src 源串 比如 "您已成功注册到%s，用户名：%s，密码：%s，请登录..."
     * @param placeholder 占位符号 比如 "%s"
     * @param replaceList 替换列表，比如 {"江苏移动","张三","111111"};
     * @return 替换后的内容，例子的内容则为： "您已成功注册到江苏移动，用户名：张三，密码：111111，请登录..."
     */
    public static String replaceSequence(String src, String placeholder,String[] replaceList)
    {
        StringBuffer buf = new StringBuffer();
        String[] segmentArray = src.split(placeholder);
        int replaceListLen = replaceList.length;
        if (segmentArray != null)
        {
            int len = segmentArray.length;
            int i = 0;
            for (; i < len - 1 && i < replaceListLen; i++)
            {
                String segment = segmentArray[i];
                buf.append(segment).append((String) replaceList[i]);
            }
            for (int j = i; j < len; j++)
                buf.append(segmentArray[j]);
        }
        return buf.toString();
    }


    /**
     * 随机生成一定长度的数字
     * 
     * @param length 长度
     * @return 字符串
     */
    public static String getRandomDigits(int length)
    {
        StringBuffer strb = new StringBuffer();
        Random random = new Random();
        for (int i=0;i<length;i++)
        {
            strb.append(random.nextInt(10));
        }
        
        return strb.toString();
    }
    
    /**
     * 随机生成一定长度的大写字母
     * 
     * @param length 长度
     * @return 字符串
     */
    public static String getRandomUpperLetters(int length)
    {
        StringBuffer strb = new StringBuffer();
        Random random = new Random();
        for (int i=0;i<length;i++)
        {
            int index = random.nextInt(26);
            strb.append(LETTERS_UPPERCASE.charAt(index));
        }
        
        return strb.toString();
    }
    
    /**
     * 随机生成一定长度的小写字母
     * 
     * @param length 长度
     * @return 字符串
     */
    public static String getRandomLowerLetters(int length)
    {
        StringBuffer strb = new StringBuffer();
        Random random = new Random();
        for (int i=0;i<length;i++)
        {
            int index = random.nextInt(26);
            strb.append(LETTERS_LOWERCASE.charAt(index));
        }
        
        return strb.toString();
    }
    
    /**
     * 随机生成一定长度的小写字母
     * 
     * @param length 长度
     * @return 字符串
     */
    public static String getRandomLowerLettersDigits(int length)
    {
        StringBuffer strb = new StringBuffer();
        Random random = new Random();
        for (int i=0;i<length;i++)
        {
            int index = random.nextInt(36);
            strb.append(LETTERS_DIGITS_LOWERCASE.charAt(index));
        }
        
        return strb.toString();
    }
    
    /**
     * 随机生成一定长度的字母
     * 
     * @param length 长度
     * @return 字符串
     */
    public static String getRandomLetters(int length)
    {
        StringBuffer strb = new StringBuffer();
        Random random = new Random();
        for (int i=0;i<length;i++)
        {
            int index = random.nextInt(52);
            strb.append(LETTERS.charAt(index));
        }
        
        return strb.toString();
    }
    
    /**
     * 随机生成一定长度的字母或数字
     * 
     * @param length 长度
     * @return 字符串
     */
    public static String getRandomLettersDigits(int length)
    {
        StringBuffer strb = new StringBuffer();
        Random random = new Random();
        for (int i=0;i<length;i++)
        {
            int index = random.nextInt(62);
            strb.append(LETTERS_DIGITS.charAt(index));
        }
        
        return strb.toString();
    }
    
    /**
     * 随机生成一定长度的字符或数字
     * 
     * @param radomLength 长度
     * @param type 类型表名生成的随机字符串是字母数字(0),数字(1),字母(2),大写字母(3),小写字母(4),大写字母和数字(5),小写字母和数字(6)
     * @return 字符串
     */
    public static String getRandomValue(int randomLength, int type)
    {
        if (randomLength < 1)
            return "";
        
        int maxInt = 62;
        String sendString = LETTERS_DIGITS;
        switch (type)
        {
        case 1:
            maxInt = 10;
            sendString = DIGITS;
            break;
        case 2:
            maxInt = 52;
            sendString = LETTERS;
            break;
        case 3:
            maxInt = 26;
            sendString = LETTERS_UPPERCASE;
            break;
        case 4:
            maxInt = 26;
            sendString = LETTERS_LOWERCASE;
            break;
        case 5:
            maxInt = 36;
            sendString = LETTERS_UPPERCASE + DIGITS;
            break;
        case 6:
            maxInt = 36;
            sendString = LETTERS_LOWERCASE + DIGITS;
            break;
        }
        
        StringBuffer returnString = new StringBuffer();
        Random random = new Random();
        
        for (int i=0;i<randomLength;i++)
        {
            int rand = random.nextInt(maxInt);
            returnString.append(sendString.charAt(rand));
        }
        
        return returnString.toString();
    }

    /*********************************/
    //以下为字符串截取相关
    /*********************************/
    
    /** 获取一个字符在字符串出现的次数 */
    public static int getTimes(String src, char c)
    {
    	int times = 0;
        for (int i=0;i<src.length();i++)
        {
        	char curc = (char)src.charAt(i);
            if (curc == c)
                times++;
        }
        
        return times;
    }

    /** 获取一个字符串在另字符串出现的次数 */
    public static int getTimes(String src, String timeStr)
    {
        int times = 0;
        if (src == null || timeStr == null)
            return times;
        
        int len = src.length();
        int timelen = timeStr.length();
        if (len < timelen)
            return times;
        
        for (int i=src.indexOf(timeStr,len-src.length());i!=-1;i=src.indexOf(timeStr,len-src.length()))
        {
            times++;
            src = src.substring(i+timelen);
        }
        
        return times;
    }

    
    /**
     * 根据srcPage来得到配对的returnPage相对路径 在WEBActionServlet用到了，很重要 
     * 
     * @param returnPage 绝对路径或相对路径
     * @param srcPage 绝对路径
     * @return 配对后的相对路径
     */
    public static String convertPath(String returnPage,String srcPage)
    {
        //如果不是/开头，则认为是相对路径，返回
        if (!returnPage.startsWith("/"))
            return returnPage;
        
        //绝对路径,去除"/"
        returnPage = returnPage.substring(1);
        
        int count = StringUtil.getTimes(srcPage,'/');
        if (count > 1)
        {
            for (int i=1;i<count;i++)
                returnPage = "../"+returnPage;
        }
        
        return returnPage;
    }
    
    /** 删除s中所有空白 (包括空格,\t,\r,\n) */
    public static String removeWhitespace(String s)
    {
        return removeCharsInBag(s, WHITE_SPACE);
    }

    /** 删除字符串前面的空白,直到出现内容 */
    public static String removeInitialWhitespace(String s)
    {
        int i = 0;
        while ((i < s.length()) && ValidateUtil.isCharInString(s.charAt(i), WHITE_SPACE))
            i++;
        return s.substring(i);
    }
    
    /** 
     * 删除s中出现的所有bag定义的字符
     * 
     * 举例: s = "adddsg"; bag = "ds"; 得到结果是:returnString = "ag";
     * @param s 原字符串
     * @param bag 包字符串
     * @return 删除s中出现的所有bag定义的字符后的字符串
     */
    public static String removeCharsInBag(String s, String bag)
    {
        String returnString = "";
        
        if (ValidateUtil.isEmpty(s))
            return returnString;

        // 逐个字符检查,如果该字符不在bag中,则加到returnString中
        for (int i=0;i<s.length();i++)
        {
            char c = s.charAt(i);

            if (bag.indexOf(c) == -1)
                returnString += c;
        }
        return returnString;
    }

    /** 
     * 删除s中所有bag未定义的字符
     * 
     * 举例: s = "adddsg"; bag = "ds"; 得到结果是:returnString = "ddds";
     * @param s 原字符串
     * @param bag 包字符串
     * @return 删除s中出现的所有bag未定义的字符后的字符串
     */
    public static String removeCharsNotInBag(String s, String bag)
    {
        String returnString = "";

        // 逐个字符检查,如果该字符在bag中,则加到returnString中
        for (int i=0;i<s.length();i++)
        {
            char c = s.charAt(i);

            if (bag.indexOf(c) != -1)
                returnString += c;
        }
        return returnString;
    }
    
    /*********************************/
    //以下为编码相关
    /*********************************/
    
    /**
     * 中文(GBK,GB2312)编码到UTF8 适用于WAP &#x;编码
     * 
     * @param sChinese
     * @return String UTF8编码
     */
    public static String toWAPUTF8(String sChinese)
    {
        if(ValidateUtil.isEmpty(sChinese)) 
            return "";
        
        String retStr = "";
        String tempStr = "";
        for(int i=0;i<sChinese.length();i++)
        {
        	tempStr = "&#x"+Integer.toHexString((int)sChinese.charAt(i)) + ";";
        	retStr += tempStr;
        }
        
        return retStr;
    }
    
    /**
     * 中文(GBK,GB2312双字节)编码到Unicode 适用于Application \\u编码
     * 
     * @param sChinese
     * @return
     */
    public static String toUnicode(String sChinese)
    {
        if(ValidateUtil.isEmpty(sChinese)) 
            return "";
        
        String retStr = "";
        String tempStr = "";
        for (int i = 0; i < sChinese.length(); i++)
        {
            tempStr = "\\u" + Integer.toHexString((int) sChinese.charAt(i));
            retStr += tempStr;
        }
        return retStr;
    }
    
    /** 4节字 to ip */
    public static String byteToIp(byte[] value)
    {
        StringBuffer strb = new StringBuffer();
        strb.append(value[0] & 0xFF).append(".");
        strb.append(value[1] & 0xFF).append(".");
        strb.append(value[2] & 0xFF).append(".");
        strb.append(value[3] & 0xFF);
        return strb.toString();
    }
    
    /** ip to long */
    public static long ipToLong(String ip)
    {
        String[] strs = ip.split("\\.");
        int[] ints = toIntArray(strs);
        int ipInt = (ints[0] << 24) + (ints[1] << 16) + (ints[2] << 8) + (ints[3]);   
        long ipLong = ipInt & 0x7FFFFFFFL;
        if (ipLong < 0)
            ipLong |= 0x80000000L;
        
        return ipLong;
    }
    
    /*********************************/
    //以下为文件相关
    /*********************************/

	/**
	 * 取得系统的文件分隔符
	 * 
	 * @return 系统分隔符，在window系统返回"\"，在unix/linux系统返回"/"
	 */
	public static String getSystemSeparator()
	{
		return System.getProperty("file.separator");
	}
	
	/**
	 * 检验路径 主要是滤去多余的分隔符 在WIN中
	 * 
	 * @param path 要检验路径
	 * @return 检验后的路径
	 */
	public static String formatWinPath(String path)
	{
		path = path.replaceAll("//", "/");
		return path;
	}
	
	/**
	 * 取得URL或文件后缀 注:这里只是找"." 
	 * 
	 * @param path URL或文件名
	 * @return 后缀,小写
	 */
	public static String getPathSuffix(String path)
	{
		if (ValidateUtil.isEmpty(path))
			return "";

		int pos = path.lastIndexOf(".");
		String fileExt = path.substring(pos + 1, path.length());
		return fileExt.toLowerCase();
	}
	
	/**
	 * 取得文件类型,则取得文件后缀 注:这里只是找"."
	 * 
	 * @param fileName 文件名
	 * @return 文件后缀,小写
	 */
	public static String getFileExt(String fileName)
	{
		if (ValidateUtil.isEmpty(fileName))
			return "";

		int pos = fileName.lastIndexOf(".");
		String fileExt = fileName.substring(pos + 1, fileName.length());
		return fileExt.toLowerCase();
	}

	/**
     * 提取一个文件路径的目录结构， 如c:\\temp\\article.jsp，则返回c:\\temp\\。函数主要 用于提取路径的目录部份
     * 
     * @param filePath 文件完整路径
     * @return 目录结构
     */
    public static String getFilePath(String filePath)
    {
        return getFilePath(filePath, getSystemSeparator());
    }
    
	/**
	 * 提取一个文件路径的目录结构， 如c:\\temp\\article.jsp，则返回c:\\temp\\。函数主要 用于提取路径的目录部份
	 * 
	 * @param filePath 文件完整路径
	 * @param sep 分隔符
	 * @return 目录结构
	 */
	public static String getFilePath(String filePath, String sep)
	{
		int pos = filePath.lastIndexOf(sep);
		if (pos == -1)
			return "";
		
		return filePath.substring(0, pos);
	}

	/**
	 * 提取一个文件路径的目录结构， 如c:\\temp\\article.jsp，则返回c:\\temp\\。函数主要 用于提取路径的目录部份
	 * 
	 * @param filePath 文件完整路径
	 * @return 目录结构
	 */
	public static String getFileURL(String filePath)
	{
		int pos = filePath.lastIndexOf("/");
		if (pos == -1)
			return "";
		return filePath.substring(0, pos);
	}

	/**
	 * 提取一个文件路径的目录结构， 如c:\\temp\\article.jsp，则返回article.jsp。函数主要 用于提取路径的文件名称
	 * 
	 * @param filePath 文件完整路径
	 * @return 文件名称
	 */
	public static String getFileName(String filePath)
	{
		return getFileName(filePath, getSystemSeparator());
	}
	
	/**
     * 提取一个文件路径的目录结构， 如c:\\temp\\article.jsp，则返回article.jsp。函数主要 用于提取路径的文件名称
     * 
     * @param filePath 文件完整路径
     * @param sep 分隔符
     * @return 文件名称
     */
    public static String getFileName(String filePath, String sep)
    {
        int pos = filePath.lastIndexOf(sep);
        if (pos == -1)
            return "";
        return filePath.substring(pos + 1);
    }
    
	/**
	 * 提取目录路径，以当前时间组成一个路径
	 * 
	 * @return String 完整路径
	 */
	public static String getPathByCurrentDate()
	{
		Calendar date = Calendar.getInstance();
		int year = date.get(Calendar.YEAR);
		int month = date.get(Calendar.MONTH) + 1;
		int day = date.get(Calendar.DATE);

		return year + getSystemSeparator() + month + getSystemSeparator() + day;
	}

	/**
	 * 提取目录路径，以参考时间组成一个完整路径
	 * @return String 完整路径
	 */
	public static String getPathByCurrentDate(Date date)
	{
		Calendar calender = Calendar.getInstance();
		calender.setTime(date);

		int year = calender.get(Calendar.YEAR);
		int month = calender.get(Calendar.MONTH) + 1;
		int day = calender.get(Calendar.DATE);

		return year + getSystemSeparator() + month + getSystemSeparator() + day;

	}
	  
    /**
     * 通过给定初始str,和前置char,得到给定长度的值, 通常用于前补0等
     * 
     * @param str 初始str
     * @param len 给定长度
     * @param prefix 前置char
     * @return String
     */
    public static String getPrefixFixLenStr(String str, int len, char prefix)
    {
        String prefixStr = "";
        for (int i=0;i<len;i++)
        {
            prefixStr += prefix;
        }
        
        str = prefix + str;
        return str.substring(str.length() - len);
    }
    
    /**
     * 通过给定初始int,和前置char,得到给定长度的值, 通常用于前补0等
     * 
     * @param str 初始str
     * @param len 给定长度
     * @param prefix 前置char
     * @return String
     */
    public static String getPrefixFixLenStr(int intValue, int len, char prefix)
    {
        String str = "";
        for (int i=0;i<len;i++)
        {
            str += prefix;
        }
        
        str = str + intValue;
        return str.substring(str.length() - len);
    }
    
    /**
     * 比较两个字符串的编码大小,默认GBK编码
     * 
     * @param s1 第一个字符串
     * @param s2 第二个字符串
     * @return 如果第一个比第二个在字典(内码)前,则<0,否则>0
     */
    public static int compare(String s1, String s2)
    {
        String m_s1 = null, m_s2 = null;
        try
        { // 先将两字符串编码成GBK
            m_s1 = new String(s1.getBytes(), "GBK");
            m_s2 = new String(s2.getBytes(), "GBK");
        }
        catch (Exception ex)
        {
            return s1.compareTo(s2);
        }

        return chineseCompareTo(m_s1, m_s2);
    }

    /**
     * 比较两个字符串的编码大小, 取机器编码
     * 
     * @param s1 第一个字符串
     * @param s2 第二个字符串
     * @return 如果第一个比第二个在字典(内码)前,则<0,否则>0
     */
    public static int chineseCompareTo(String s1, String s2)
    {
        int len1 = s1.length();
        int len2 = s2.length();
        int n = Math.min(len1, len2);
        for (int i = 0; i < n; i++)
        {
            int s1_code = getCharCode(s1.charAt(i) + "");
            int s2_code = getCharCode(s2.charAt(i) + "");
            if (s1_code != s2_code)
                return s1_code - s2_code;
        }
        return len1 - len2;
    }
    
    /**
     * 通过一个字符串,读取他的CHAR,保证第一个字符是汉字或英文(取两位)
     * 
     * @param s 字符串
     * @return 对应int型
     */
    public static int getCharCode(String s)
    {
        if (s == null || s.length() == 0)
            return -1;

        byte[] b = s.getBytes();
        int value = 0; // 保证取第一个字符（汉字或者英文）
        for (int i = 0; i < b.length && i <= 2; i++)
        {
            value = value * 100 + b[i];
        }

        return value;
    }
    
    /**
     * 提供被除数，除数和小数位数，得到结果
     * 
     * @param dividend 被除数
     * @param divisor 除数
     * @param radixLen 小数位数
     * @return String 结果
     */
    public static String getDivsionString(long dividend, int divisor, int radixLen)
    {
        double result = (double)dividend / divisor;
        StringBuffer radix = new StringBuffer("#");
        if (radixLen > 0)
        {
            radix.append(".");
            for (int i=0;i<radixLen;i++)
                radix.append("#");
        }
        
        DecimalFormat df = new DecimalFormat(radix.toString());
        String ret = df.format(result);
        
        if (radixLen > 0)
        {
            int ind = ret.indexOf('.');
            if (ind == -1)
            {//没有小数点
                ret += ".";
                for (int i=0;i<radixLen;i++)
                    ret += "0";
            }
            else if (ind > ret.length() - radixLen -1)
            {//小数位数不足,尾部加0
                int zeroNum = ind - (ret.length() - radixLen - 1);
                for (int i=0;i<zeroNum;i++)
                {
                    ret += "0";
                }
            }
        }
        
        return ret;
    }
    
    public static String decimal2Chinese(int value) 
    {
        Integer in = new Integer(value);
        double src = in.doubleValue();
        src = src/100;
        
        StringBuilder sb = new StringBuilder();
        DecimalFormat df = new DecimalFormat("0.00");
        String srcText = df.format(src);
        String numText = srcText.substring(0, srcText.length() - 3);
        String decText = srcText.substring(srcText.length() - 2);

        int numlen = numText.length();
        for (int k = 0; k < numlen; k++) 
        {
            sb.append(RMB_NUM[Integer.parseInt(String.valueOf(srcText.charAt(k)))]);
            sb.append(RMB_UNIT[numlen - k - 1]);
        }
        if ("00".equals(decText))
        {
            sb.append("整");
        } 
        else
        {
            sb.append(RMB_NUM[Integer.parseInt(String.valueOf(decText.charAt(0)))]);
            sb.append(RMB_DEC[0]);
            sb.append(RMB_NUM[Integer.parseInt(String.valueOf(decText.charAt(1)))]);
            sb.append(RMB_DEC[1]);
        }
        String result = sb.toString();
        result = result.replace("零分", "");
        result = result.replace("零角", "零");
        result = result.replace("零仟", "零");
        result = result.replace("零佰", "零");
        result = result.replace("零拾", "零");
        result = result.replace("零圆", "圆");
        while (true)
        {
            String r = result.replace("零零", "零");
            if (r.equals(result))
            {
                break;
            } 
            else 
            {
                result = r;
            }
        }
        result = result.replace("零圆", "圆");
        result = result.replace("零万", "万");
        result = result.replace("零亿", "亿");
        result = result.replace("亿万", "亿");
        if(result.startsWith("圆"))
        {
            result="零"+result;
        }
        return result;
    }

    /**
     * 读取资源文件 path格式为/com/zoulab/res/abc.js
     * 
     * @param path 路径
     * @return String
     * @throws IOException
     */
    public static String readResource(String path) throws IOException
    {
        return readResource(StringUtil.class, path);
    }
    
    /**
     * 读取资源文件 path格式为/com/zoulab/res/abc.js
     * 
     * @param clazz 类名
     * @param path 路径
     * @return String
     * @throws IOException
     */
    public static String readResource(Class<?> clazz, String path) throws IOException
    {
        InputStream in = null;
        byte[] buf = null;
        
        in = clazz.getResourceAsStream(path);
        buf = new byte[in.available()];
        in.read(buf);
        in.close();
        in = null;
        
        String res = new String(buf);
        res = res.replaceAll("\r\n"," ");
        res = res.replaceAll("\r", " ");
        res = res.replaceAll("\n", " ");
        return res;
    }
    
    /**
     * 快速转换成小写，只对ASCII格式，UNICODE不转换
     * 
     * @param 原字符串
     * @return 目标字符串
     */
    public static String asciiToLowerCase(String s)
    {
        char[] c = null;
        int i=s.length();

        // look for first conversion
        while (i-->0)
        {
            char c1=s.charAt(i);
            if (c1<=127)
            {
                char c2=LOWER_CASES[c1];
                if (c1!=c2)
                {
                    c=s.toCharArray();
                    c[i]=c2;
                    break;
                }
            }
        }

        while (i-->0)
        {
            if(c[i]<=127)
                c[i] = LOWER_CASES[c[i]];
        }
        
        return c==null?s:new String(c);
    }


    /**
     * 忽略大小写验证startsWith
     * 
     * @param s 被验证字符串
     * @param w 验证字符串
     * @return boolean =true表示startsWith, 否则=false
     */
    public static boolean startsWithIgnoreCase(String s,String w)
    {
        if (w==null)
            return true;
        
        if (s==null || s.length()<w.length())
            return false;
        
        for (int i=0;i<w.length();i++)
        {
            char c1=s.charAt(i);
            char c2=w.charAt(i);
            if (c1!=c2)
            {
                if (c1<=127)
                    c1=LOWER_CASES[c1];
                if (c2<=127)
                    c2=LOWER_CASES[c2];
                if (c1!=c2)
                    return false;
            }
        }
        return true;
    }
    
    /**
     * 忽略大小写验证endsWith
     * 
     * @param s 被验证字符串
     * @param w 验证字符串
     * @return boolean =true表示endsWith, 否则=false
     */
    public static boolean endsWithIgnoreCase(String s,String w)
    {
        if (w==null)
            return true;
        
        int sl=s.length();
        int wl=w.length();
        
        if (s==null || sl<wl)
            return false;
        
        for (int i=wl;i-->0;)
        {
            char c1=s.charAt(--sl);
            char c2=w.charAt(i);
            if (c1!=c2)
            {
                if (c1<=127)
                    c1=LOWER_CASES[c1];
                if (c2<=127)
                    c2=LOWER_CASES[c2];
                if (c1!=c2)
                    return false;
            }
        }
        return true;
    }
    
    /**
     * 金额字符串转int金额分，支持两位小数点的金额字符串
     * @param str  金额字符串
     * @param defaultValue 缺省值
     * @return int金额分
     */
    public static int getMoneyTwoRadix(String str, int defaultValue)
    {
        if (str == null)
            return defaultValue;
        
        if (!ValidateUtil.isMoneyTwoRadix(str))
            return defaultValue;
        
        boolean isNegative = false;
        if (str.startsWith("-"))
        {
            isNegative = true;
            str = str.substring(1);
        }
        
        int index = str.indexOf('.');
        if (index == -1)
        {
            int value = Integer.parseInt(str) * 100;//由元转为分
            return (isNegative)?-value:value;
        }
        
        int integer = Integer.parseInt(str.substring(0, index)) * 100;
        int radix = Integer.parseInt(str.substring(index + 1));
        int value = integer + radix;
        return (isNegative)?-value:value;
    }



}
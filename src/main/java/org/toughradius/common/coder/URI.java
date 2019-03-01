package org.toughradius.common.coder;


import org.toughradius.common.ValidateUtil;

/**
 * URI [scheme://host[:port]]/path[;jsessionid][?query]
 */
public class URI
{
    private static final String SESSION_KEY = ";jsessionid=";
    private static final String DOMAIN  = ".:/"
                                        + "0123456789"
                                        + "abcdefghijklmnopqrstuvwxyz"
                                        + "ABCDEFGHIJKLMNOPQRSTUVWXYZ";
    private static final String SESSION = "?"
                                        + "0123456789"
                                        + "abcdefghijklmnopqrstuvwxyz"
                                        + "ABCDEFGHIJKLMNOPQRSTUVWXYZ";
    
    private String uri;
    private String scheme = "http";
    private String host;
    private int port;
    private String path;
    private String sessionId;
    private String query;
    
    private String pathInContext;
    
    public String toString()
    {
        return uri;
    }
    
    public String getScheme()
    {
        return scheme;
    }
    
    public String getHost()
    {
        return host;
    }
    
    public int getPort()
    {
        return port;
    }
    
    public String getPath()
    {
        return path;
    }
    
    public String getSessionId()
    {
        return sessionId;
    }
    
    public String getQuery()
    {
        return query;
    }

    public String getPathInContext()
    {
        return pathInContext;
    }

    public void setPathInContext(String pathInContext)
    {
        this.pathInContext = pathInContext;
    }

    /** 从path中解释虚拟目录 */
    public String getVirtualDirectory()
    {
        int ind = path.indexOf('/', 1);
        if (ind != -1)
            return path.substring(0, ind);//虚拟目录 如 /abc/tests.html
        else
            return "/";//根目录, 如 /abc.html 
    }
    
    /**
     * 解析REQUEST-URI信息
     * @param uri [scheme://host[:port]]/path[;jsessionid][?query]
     * @return =true 表示成功，=false 表示失败
     */
    public boolean parseUri(String uri)
    {
        this.uri = uri;
        
        boolean hasHost = true;
        if (uri.startsWith("http://"))
            uri = uri.substring(7);
        else if (uri.startsWith("https://"))
            uri = uri.substring(8);
        else
        {//相对路径/开头,未提供host
            hasHost = false;
            port = 80;
        }
        
        if (hasHost)
        {//读取绝对路径时提供的host
            int i = 0;
            StringBuffer strb = new StringBuffer();
            for (;i<uri.length();i++)
            {
                char c = uri.charAt(i);
                if (!ValidateUtil.isCharInString(c, DOMAIN))
                    return false;

                if (c == ':')
                    break;//host结束
                else if (c != '/')
                    strb.append(c);//host未结束
                else
                {//host结束，且默认端口
                    port = 80;
                    break;
                }
            }
            
            if (strb.length() == 0)
                return false;
            else
            {
                host = strb.toString().toLowerCase();
                uri = uri.substring(i);
                if (uri.length() <= 0)
                    port = 80;//uri结尾
            }
        }
        
        if (port == 0 && uri.length() > 0)
        {
            if (!uri.startsWith(":"))
                return false;
            
            uri = uri.substring(1);
            int i = 0;
            StringBuffer strb = new StringBuffer();
            for (;i<uri.length();i++)
            {
                char c = uri.charAt(i);
                if (!ValidateUtil.isCharInString(c, "/0123456789"))
                    return false;

                if (c == '/')
                    break;//port结束
                else
                    strb.append(c);//port未结束
            }
            
            if (strb.length() == 0)
                return false;
            else
            {
                port = Integer.parseInt(strb.toString());
                uri = uri.substring(i);
            }
        }
        
        boolean hasSessionId = false;
        if (uri.length() > 0)
        {
            if (!uri.startsWith("/"))
                return false;
            
            int i = 0;
            StringBuffer strb = new StringBuffer();
            for (;i<uri.length();i++)
            {
                char c = uri.charAt(i);
                if (c == '?')
                    break;//path结束
                else if (c != ';')
                    strb.append(c);//path未结束
                else
                {
                    hasSessionId = true;//path结束,后有sessionId
                    break;
                }
            }
            
            if (strb.length() > 0)
            {
                path = strb.toString();
                uri = uri.substring(i);
            }
        }
        
        if (hasSessionId && uri.length() > 0)
        {
            if (!uri.startsWith(SESSION_KEY))
                return false;
            
            uri = uri.substring(SESSION_KEY.length());
            
            int i = 0;
            StringBuffer strb = new StringBuffer();
            for (;i<uri.length();i++)
            {
                char c = uri.charAt(i);
                if (!ValidateUtil.isCharInString(c, SESSION))
                    return false;

                if (c == '?')
                    break;//port结束
                else
                    strb.append(c);//port未结束
            }
            
            if (strb.length() > 0)
            {
                sessionId = strb.toString();
                uri = uri.substring(i);
            }
        }
        
        if (uri.length() > 0)
        {
            if (!uri.startsWith("?"))
                return false;
            
            query = uri.substring(1);
        }
        
        return true;
    }
    
    public static void main(String[] args)
    {
        URI uri = new URI();
        uri.parseUri("/");
        System.out.println(uri.getScheme());
        System.out.println(uri.getHost());
        System.out.println(uri.getPort());
        System.out.println(uri.getPath());
        System.out.println(uri.getSessionId());
        System.out.println(uri.getQuery());
    }
}

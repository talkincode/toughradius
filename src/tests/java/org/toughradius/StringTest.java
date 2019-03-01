package org.toughradius;

import org.junit.Test;

import java.io.UnsupportedEncodingException;

public class StringTest {

    @Test
    public void testSize() throws UnsupportedEncodingException {
        String s = "在网络层，因为IP包的首部要占用20字节，所以这的MTU为1500－20＝1480；　\n" +
                "3.在传输层，对于UDP包的首部要占用8字节，所以这的MTU为1480－8＝1472； 　　\n" +
                "所以，在应用层，你的Data最大长度为1472。当我们的UDP包中的数据多于MTU(1472)时，" +
                "发送方的IP层需要分片fragmentation进行传输，而在接收方IP层则需要进行数据报重组，" +
                "由于UDP是不可靠的传输协议，如果分片丢失导致重组失败，将导致UDP数据包被丢弃。 　";
        System.out.println(s.getBytes("UTF-8").length);
    }
}

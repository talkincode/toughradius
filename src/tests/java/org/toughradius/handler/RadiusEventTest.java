package org.toughradius.handler;

import org.junit.Test;

import java.net.*;
import java.nio.ByteBuffer;

public class RadiusEventTest {

    @Test
    public void testSendEvent() throws Exception {
        DatagramPacket packetIn = new DatagramPacket(new byte[1], 1);
        DatagramSocket socket = new DatagramSocket();
        byte[] username = "test01".getBytes("utf-8");
        byte[] content = "hello world".getBytes("utf-8");
        int length = 1+2+username.length+2+content.length;
        ByteBuffer buff = ByteBuffer.allocate(length );
        buff.put((byte)0xe);
        buff.putShort((short)username.length);
        buff.put(username);
        buff.putShort((short)content.length);
        buff.put(content);
        buff.flip();
        byte[] data = new byte[length];
        buff.get(data);

        socket.send(new DatagramPacket(data,data.length,InetAddress.getByName("127.0.0.1"),1814));
        socket.receive(packetIn);
        assert  packetIn.getData()[0] == 0x00;
    }
}

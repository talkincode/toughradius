package org.toughradius.portal;

import org.apache.mina.core.buffer.IoBuffer;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.stereotype.Component;
import org.toughradius.portal.packet.PortalPacket;

import java.io.IOException;
import java.net.DatagramPacket;
import java.net.DatagramSocket;
import java.net.InetSocketAddress;
import java.net.SocketException;

@Component
@ConfigurationProperties(prefix = "xspeeder.cloud.portal")
public class PortalClient {

    private int trace;
    private int timeout = 5000;

    public int getTrace() {
        return trace;
    }

    public void setTrace(int trace) {
        this.trace = trace;
    }

    public int getTimeout() {
        return timeout;
    }

    public void setTimeout(int timeout) {
        this.timeout = timeout;
    }

    public PortalPacket sendToAc(PortalPacket request, String ipaddr, int port) throws PortalException {
        DatagramSocket sock = null;
        try{
            sock = new DatagramSocket();
            IoBuffer buff = request.encodePacket();
            byte[] data = new byte[buff.remaining()];
            buff.get(data);
            DatagramPacket packetOut = new DatagramPacket(data,data.length, new InetSocketAddress(ipaddr,port));
            DatagramPacket packetIn = new DatagramPacket(new byte[PortalPacket.MAX_PACKET_LENGTH],PortalPacket.MAX_PACKET_LENGTH);
            sock.setSoTimeout(timeout);
            sock.send(packetOut);
            sock.receive(packetIn);
            PortalPacket resp = new PortalPacket(packetIn.getData());
            resp.checkResponseAuthenticator(request.getSecret(),request.getAuthenticator());
            return resp;
        } catch (SocketException e) {
            throw new PortalException("网络异常",e);
        } catch (IOException e) {
            throw new PortalException("IO异常",e);
        }
    }

    public void sendToAcNoReply(PortalPacket request,String ipaddr, int port) throws PortalException {
        DatagramSocket sock = null;
        try{
            sock = new DatagramSocket();
            IoBuffer buff = request.encodePacket();
            byte[] data = new byte[buff.remaining()];
            buff.get(data);
            DatagramPacket packetOut = new DatagramPacket(data,data.length, new InetSocketAddress(ipaddr,port));
            sock.setSoTimeout(timeout);
            sock.send(packetOut);
        } catch (SocketException e) {
            throw new PortalException("网络异常",e);
        } catch (IOException e) {
            throw new PortalException("IO异常",e);
        }finally{
            try{
                sock.close();
            } catch(Exception e){
            }
        }
    }
}

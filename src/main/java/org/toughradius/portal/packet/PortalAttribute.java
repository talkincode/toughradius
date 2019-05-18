package org.toughradius.portal.packet;

import org.toughradius.portal.PortalException;
import org.toughradius.portal.utils.PortalUtils;

public class PortalAttribute {

    private int attributeType = -1;
    private byte[] attributeData = null;

    public PortalAttribute() {
    }

    public PortalAttribute(int attributeType, byte[] attributeData) {
        this.attributeType = attributeType;
        this.attributeData = attributeData;
    }

    public int getAttributeType() {
        return attributeType;
    }

    public void setAttributeType(int attributeType) {
        if (attributeType < 0 || attributeType > 255)
            throw new IllegalArgumentException("attribute type invalid: " + attributeType);
        this.attributeType = attributeType;
    }

    public byte[] getAttributeData() {
        return attributeData;
    }
    public void setAttributeData(byte[] attributeData) {
        if (attributeData == null)
            throw new NullPointerException("attribute data is null");
        this.attributeData = attributeData;
    }

    public void setAttributeValue(String value) {
        throw new RuntimeException("cannot set the value of attribute " + attributeType + " as a string");
    }

    public String getAttributeValue() {
        return PortalUtils.getHexString(getAttributeData());
    }

    public String getAttributeAsStr() {
        return PortalUtils.decodeString(getAttributeData());
    }

    public byte[] encodeAttribute() {
        if (getAttributeType() == -1)
            throw new IllegalArgumentException("attribute type not set");
        if (attributeData == null)
            throw new NullPointerException("attribute data not set");

        byte[] attr = new byte[2 + attributeData.length];
        attr[0] = (byte)getAttributeType();
        attr[1] = (byte) (2 + attributeData.length);
        System.arraycopy(attributeData, 0, attr, 2, attributeData.length);
        return attr;
    }

    public void decodeAttribute(byte[] data, int offset, int length) throws PortalException {
        if (length < 2)
            throw new PortalException("attribute length too small: " + length);
        int attrType = data[offset] & 0x0ff;
        int attrLen = data[offset + 1] & 0x0ff;
        byte[] attrData = new byte[attrLen - 2];
        System.arraycopy(data, offset + 2, attrData, 0, attrLen - 2);
        setAttributeType((byte)attrType);
        setAttributeData(attrData);
    }

    protected String getTypeName(int type){
        switch (type){
            case PortalPacket.ATTRIBUTE_BASIP_TYPE: return "BasIp";
            case PortalPacket.ATTRIBUTE_MAC_TYPE: return "Mac";
            case PortalPacket.ATTRIBUTE_CHAP_PWD_TYPE: return "ChapPassword";
            case PortalPacket.ATTRIBUTE_PASSWORD_TYPE: return "Password";
            case PortalPacket.ATTRIBUTE_USERNAME_TYPE: return "Username";
            case PortalPacket.ATTRIBUTE_CHALLENGE_TYPE: return "Challenge";
            case PortalPacket.ATTRIBUTE_PORT_TYPE: return "Port";
            case PortalPacket.ATTRIBUTE_TEXT_INFO_TYPE: return "Textinfo";
            default:return String.format("Unknow Type (%s)", type);
        }
    }

    public String toString() {
        String fmt = "%s (len=%s):%s";
        int attrlen = attributeData.length+2;
        switch (getAttributeType()){
            case PortalPacket.ATTRIBUTE_BASIP_TYPE:
                return String.format(fmt, getTypeName(attributeType),attrlen, PortalUtils.decodeIpv4(getAttributeData()));
            case PortalPacket.ATTRIBUTE_MAC_TYPE:
                return String.format(fmt, getTypeName(attributeType),attrlen, PortalUtils.decodeMacAddr(getAttributeData()));
            case PortalPacket.ATTRIBUTE_CHAP_PWD_TYPE:
                return String.format(fmt, getTypeName(attributeType),attrlen,getAttributeValue());
            case PortalPacket.ATTRIBUTE_PASSWORD_TYPE:
                return String.format(fmt, getTypeName(attributeType),attrlen, PortalUtils.decodeString(getAttributeData()));
            case PortalPacket.ATTRIBUTE_USERNAME_TYPE:
                return String.format(fmt, getTypeName(attributeType),attrlen, PortalUtils.decodeString(getAttributeData()));
            case PortalPacket.ATTRIBUTE_CHALLENGE_TYPE:
                return String.format(fmt, getTypeName(attributeType),attrlen, getAttributeValue());
            case PortalPacket.ATTRIBUTE_PORT_TYPE:
                return String.format(fmt, getTypeName(attributeType),attrlen, PortalUtils.decodeShort(getAttributeData()));
            case PortalPacket.ATTRIBUTE_TEXT_INFO_TYPE:
                return String.format(fmt, getTypeName(attributeType),attrlen, PortalUtils.decodeString(getAttributeData()));
            default:
                return String.format(fmt, getTypeName(attributeType),attrlen, getAttributeValue());
        }
    }


}

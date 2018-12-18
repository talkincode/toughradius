package org.toughradius.common.coder;

/**
 * DES加密解密
 */
public class DES
{
    private int[] encryptKeys = new int[32];
    private int[] decryptKeys = new int[32];

    private int[] tempInts = new int[2];

    public DES()
    {
    }

    public DES(byte[] key)
    {
        if (key.length == 7)
        {
            byte[] key8 = new byte[8];
            makeSMBKey(key, key8);
            setKey(key8);
        }
        else
        {
            setKey(key);
        }
    }
    
    /** 设置密钥 */
    public void setKey(byte[] key)
    {
        deskey(key, true, encryptKeys);
        deskey(key, false, decryptKeys);
    }
    
    /** 加密一个数组(长度必须是8的倍数) */
    public byte[] encrypt(byte[] clearText)
    {

        int length = clearText.length;

        if (length % 8 != 0)
        {
            System.out.println("长度必须是8的倍数");
            return null;
        }

        byte[] cipherText = new byte[length];
        int count = length / 8;

        for (int i = 0; i < count; i++)
            encrypt(clearText, i * 8, cipherText, i * 8);

        return cipherText;
    }

    /** 解密一个数组(长度必须是8的倍数) */
    public byte[] decrypt(byte[] cipherText)
    {

        int length = cipherText.length;

        if (length % 8 != 0)
        {
            System.out.println("长度必须是8的倍数");
            return null;
        }

        byte[] clearText = new byte[length];
        int count = length / 8;

        for (int i = 0; i < count; i++)
            decrypt(cipherText, i * 8, clearText, i * 8);

        return clearText;
    }

    private void deskey(byte[] keyBlock, boolean encrypting, int[] KnL)
    {

        int i, j, l, m, n;
        int[] pc1m = new int[56];
        int[] pcr = new int[56];
        int[] kn = new int[32];

        for (j = 0; j < 56; ++j)
        {
            l = pc1[j];
            m = l & 07;
            pc1m[j] = ((keyBlock[l >>> 3] & bytebit[m]) != 0) ? 1 : 0;
        }

        for (i = 0; i < 16; ++i)
        {

            if (encrypting)
                m = i << 1;
            else
                m = (15 - i) << 1;
            n = m + 1;
            kn[m] = kn[n] = 0;
            for (j = 0; j < 28; ++j)
            {
                l = j + totrot[i];
                if (l < 28)
                    pcr[j] = pc1m[l];
                else
                    pcr[j] = pc1m[l - 28];
            }
            for (j = 28; j < 56; ++j)
            {
                l = j + totrot[i];
                if (l < 56)
                    pcr[j] = pc1m[l];
                else
                    pcr[j] = pc1m[l - 28];
            }
            for (j = 0; j < 24; ++j)
            {
                if (pcr[pc2[j]] != 0)
                    kn[m] |= bigbyte[j];
                if (pcr[pc2[j + 24]] != 0)
                    kn[n] |= bigbyte[j];
            }
        }
        cookey(kn, KnL);
    }

    private void cookey(int[] raw, int KnL[])
    {
        int raw0, raw1;
        int rawi, KnLi;
        int i;

        for (i = 0, rawi = 0, KnLi = 0; i < 16; ++i)
        {
            raw0 = raw[rawi++];
            raw1 = raw[rawi++];
            KnL[KnLi] = (raw0 & 0x00fc0000) << 6;
            KnL[KnLi] |= (raw0 & 0x00000fc0) << 10;
            KnL[KnLi] |= (raw1 & 0x00fc0000) >>> 10;
            KnL[KnLi] |= (raw1 & 0x00000fc0) >>> 6;
            ++KnLi;
            KnL[KnLi] = (raw0 & 0x0003f000) << 12;
            KnL[KnLi] |= (raw0 & 0x0000003f) << 16;
            KnL[KnLi] |= (raw1 & 0x0003f000) >>> 4;
            KnL[KnLi] |= (raw1 & 0x0000003f);
            ++KnLi;
        }
    }


   
    
    /** 加密每8个字节 */
    private void encrypt(byte[] clearText, int clearOff, byte[] cipherText,
        int cipherOff)
    {

        squashBytesToInts(clearText, clearOff, tempInts, 0, 2);
        des(tempInts, tempInts, encryptKeys);
        spreadIntsToBytes(tempInts, 0, cipherText, cipherOff, 2);
    }

    /** 解密每8个字节 */
    private void decrypt(byte[] cipherText, int cipherOff, byte[] clearText,
        int clearOff)
    {

        squashBytesToInts(cipherText, cipherOff, tempInts, 0, 2);
        des(tempInts, tempInts, decryptKeys);
        spreadIntsToBytes(tempInts, 0, clearText, clearOff, 2);
    }

    private void des(int[] inInts, int[] outInts, int[] keys)
    {

        int fval, work, right, leftt;
        int round;
        int keysi = 0;

        leftt = inInts[0];
        right = inInts[1];

        work = ((leftt >>> 4) ^ right) & 0x0f0f0f0f;
        right ^= work;
        leftt ^= (work << 4);

        work = ((leftt >>> 16) ^ right) & 0x0000ffff;
        right ^= work;
        leftt ^= (work << 16);

        work = ((right >>> 2) ^ leftt) & 0x33333333;
        leftt ^= work;
        right ^= (work << 2);

        work = ((right >>> 8) ^ leftt) & 0x00ff00ff;
        leftt ^= work;
        right ^= (work << 8);
        right = (right << 1) | ((right >>> 31) & 1);

        work = (leftt ^ right) & 0xaaaaaaaa;
        leftt ^= work;
        right ^= work;
        leftt = (leftt << 1) | ((leftt >>> 31) & 1);

        for (round = 0; round < 8; ++round)
        {
            work = (right << 28) | (right >>> 4);
            work ^= keys[keysi++];
            fval = SP7[work & 0x0000003f];
            fval |= SP5[(work >>> 8) & 0x0000003f];
            fval |= SP3[(work >>> 16) & 0x0000003f];
            fval |= SP1[(work >>> 24) & 0x0000003f];
            work = right ^ keys[keysi++];
            fval |= SP8[work & 0x0000003f];
            fval |= SP6[(work >>> 8) & 0x0000003f];
            fval |= SP4[(work >>> 16) & 0x0000003f];
            fval |= SP2[(work >>> 24) & 0x0000003f];
            leftt ^= fval;
            work = (leftt << 28) | (leftt >>> 4);
            work ^= keys[keysi++];
            fval = SP7[work & 0x0000003f];
            fval |= SP5[(work >>> 8) & 0x0000003f];
            fval |= SP3[(work >>> 16) & 0x0000003f];
            fval |= SP1[(work >>> 24) & 0x0000003f];
            work = leftt ^ keys[keysi++];
            fval |= SP8[work & 0x0000003f];
            fval |= SP6[(work >>> 8) & 0x0000003f];
            fval |= SP4[(work >>> 16) & 0x0000003f];
            fval |= SP2[(work >>> 24) & 0x0000003f];
            right ^= fval;
        }

        right = (right << 31) | (right >>> 1);
        work = (leftt ^ right) & 0xaaaaaaaa;
        leftt ^= work;
        right ^= work;
        leftt = (leftt << 31) | (leftt >>> 1);
        work = ((leftt >>> 8) ^ right) & 0x00ff00ff;
        right ^= work;
        leftt ^= (work << 8);
        work = ((leftt >>> 2) ^ right) & 0x33333333;
        right ^= work;
        leftt ^= (work << 2);
        work = ((right >>> 16) ^ leftt) & 0x0000ffff;
        leftt ^= work;
        right ^= (work << 16);
        work = ((right >>> 4) ^ leftt) & 0x0f0f0f0f;
        leftt ^= work;
        right ^= (work << 4);
        outInts[0] = right;
        outInts[1] = leftt;
    }

    /** 加密一个数据块 */
    public void encrypt(byte[] clearText, byte[] cipherText)
    {
        encrypt(clearText, 0, cipherText, 0);
    }

    /** 解密一个数据块 */
    public void decrypt(byte[] cipherText, byte[] clearText)
    {
        decrypt(cipherText, 0, clearText, 0);
    }

    private static byte[] bytebit = {(byte)0x80,(byte)0x40,(byte)0x20,
        (byte) 0x10, (byte) 0x08, (byte) 0x04, (byte) 0x02, (byte) 0x01 };
    private static int[] bigbyte = { 0x800000, 0x400000, 0x200000, 0x100000,
        0x080000, 0x040000, 0x020000, 0x010000, 0x008000, 0x004000, 0x002000,
        0x001000, 0x000800, 0x000400, 0x000200, 0x000100, 0x000080, 0x000040,
        0x000020, 0x000010, 0x000008, 0x000004, 0x000002, 0x000001 };
    private static byte[] pc1 = { (byte) 56, (byte) 48, (byte) 40, (byte) 32,
        (byte) 24, (byte) 16, (byte) 8, (byte) 0, (byte) 57, (byte) 49,
        (byte) 41, (byte) 33, (byte) 25, (byte) 17, (byte) 9, (byte) 1,
        (byte) 58, (byte) 50, (byte) 42, (byte) 34, (byte) 26, (byte) 18,
        (byte) 10, (byte) 2, (byte) 59, (byte) 51, (byte) 43, (byte) 35,
        (byte) 62, (byte) 54, (byte) 46, (byte) 38, (byte) 30, (byte) 22,
        (byte) 14, (byte) 6, (byte) 61, (byte) 53, (byte) 45, (byte) 37,
        (byte) 29, (byte) 21, (byte) 13, (byte) 5, (byte) 60, (byte) 52,
        (byte) 44, (byte) 36, (byte) 28, (byte) 20, (byte) 12, (byte) 4,
        (byte) 27, (byte) 19, (byte) 11, (byte) 3 };
    private static int[] totrot = { 1, 2, 4, 6, 8, 10, 12, 14, 15, 17, 19, 21,
        23, 25, 27, 28 };

    private static byte[] pc2 = { (byte) 13, (byte) 16, (byte) 10, (byte) 23,
        (byte) 0, (byte) 4, (byte) 2, (byte) 27, (byte) 14, (byte) 5,
        (byte) 20, (byte) 9, (byte) 22, (byte) 18, (byte) 11, (byte) 3,
        (byte) 25, (byte) 7, (byte) 15, (byte) 6, (byte) 26, (byte) 19,
        (byte) 12, (byte) 1, (byte) 40, (byte) 51, (byte) 30, (byte) 36,
        (byte) 46, (byte) 54, (byte) 29, (byte) 39, (byte) 50, (byte) 44,
        (byte) 32, (byte) 47, (byte) 43, (byte) 48, (byte) 38, (byte) 55,
        (byte) 33, (byte) 52, (byte) 45, (byte) 41, (byte) 49, (byte) 35,
        (byte) 28, (byte) 31, };

    private static int[] SP1 = { 0x01010400, 0x00000000, 0x00010000,
        0x01010404, 0x01010004, 0x00010404, 0x00000004, 0x00010000, 0x00000400,
        0x01010400, 0x01010404, 0x00000400, 0x01000404, 0x01010004, 0x01000000,
        0x00000004, 0x00000404, 0x01000400, 0x01000400, 0x00010400, 0x00010400,
        0x01010000, 0x01010000, 0x01000404, 0x00010004, 0x01000004, 0x01000004,
        0x00010004, 0x00000000, 0x00000404, 0x00010404, 0x01000000, 0x00010000,
        0x01010404, 0x00000004, 0x01010000, 0x01010400, 0x01000000, 0x01000000,
        0x00000400, 0x01010004, 0x00010000, 0x00010400, 0x01000004, 0x00000400,
        0x00000004, 0x01000404, 0x00010404, 0x01010404, 0x00010004, 0x01010000,
        0x01000404, 0x01000004, 0x00000404, 0x00010404, 0x01010400, 0x00000404,
        0x01000400, 0x01000400, 0x00000000, 0x00010004, 0x00010400, 0x00000000,
        0x01010004 };
    private static int[] SP2 = { 0x80108020, 0x80008000, 0x00008000,
        0x00108020, 0x00100000, 0x00000020, 0x80100020, 0x80008020, 0x80000020,
        0x80108020, 0x80108000, 0x80000000, 0x80008000, 0x00100000, 0x00000020,
        0x80100020, 0x00108000, 0x00100020, 0x80008020, 0x00000000, 0x80000000,
        0x00008000, 0x00108020, 0x80100000, 0x00100020, 0x80000020, 0x00000000,
        0x00108000, 0x00008020, 0x80108000, 0x80100000, 0x00008020, 0x00000000,
        0x00108020, 0x80100020, 0x00100000, 0x80008020, 0x80100000, 0x80108000,
        0x00008000, 0x80100000, 0x80008000, 0x00000020, 0x80108020, 0x00108020,
        0x00000020, 0x00008000, 0x80000000, 0x00008020, 0x80108000, 0x00100000,
        0x80000020, 0x00100020, 0x80008020, 0x80000020, 0x00100020, 0x00108000,
        0x00000000, 0x80008000, 0x00008020, 0x80000000, 0x80100020, 0x80108020,
        0x00108000 };
    private static int[] SP3 = { 0x00000208, 0x08020200, 0x00000000,
        0x08020008, 0x08000200, 0x00000000, 0x00020208, 0x08000200, 0x00020008,
        0x08000008, 0x08000008, 0x00020000, 0x08020208, 0x00020008, 0x08020000,
        0x00000208, 0x08000000, 0x00000008, 0x08020200, 0x00000200, 0x00020200,
        0x08020000, 0x08020008, 0x00020208, 0x08000208, 0x00020200, 0x00020000,
        0x08000208, 0x00000008, 0x08020208, 0x00000200, 0x08000000, 0x08020200,
        0x08000000, 0x00020008, 0x00000208, 0x00020000, 0x08020200, 0x08000200,
        0x00000000, 0x00000200, 0x00020008, 0x08020208, 0x08000200, 0x08000008,
        0x00000200, 0x00000000, 0x08020008, 0x08000208, 0x00020000, 0x08000000,
        0x08020208, 0x00000008, 0x00020208, 0x00020200, 0x08000008, 0x08020000,
        0x08000208, 0x00000208, 0x08020000, 0x00020208, 0x00000008, 0x08020008,
        0x00020200 };
    private static int[] SP4 = { 0x00802001, 0x00002081, 0x00002081,
        0x00000080, 0x00802080, 0x00800081, 0x00800001, 0x00002001, 0x00000000,
        0x00802000, 0x00802000, 0x00802081, 0x00000081, 0x00000000, 0x00800080,
        0x00800001, 0x00000001, 0x00002000, 0x00800000, 0x00802001, 0x00000080,
        0x00800000, 0x00002001, 0x00002080, 0x00800081, 0x00000001, 0x00002080,
        0x00800080, 0x00002000, 0x00802080, 0x00802081, 0x00000081, 0x00800080,
        0x00800001, 0x00802000, 0x00802081, 0x00000081, 0x00000000, 0x00000000,
        0x00802000, 0x00002080, 0x00800080, 0x00800081, 0x00000001, 0x00802001,
        0x00002081, 0x00002081, 0x00000080, 0x00802081, 0x00000081, 0x00000001,
        0x00002000, 0x00800001, 0x00002001, 0x00802080, 0x00800081, 0x00002001,
        0x00002080, 0x00800000, 0x00802001, 0x00000080, 0x00800000, 0x00002000,
        0x00802080 };
    private static int[] SP5 = { 0x00000100, 0x02080100, 0x02080000,
        0x42000100, 0x00080000, 0x00000100, 0x40000000, 0x02080000, 0x40080100,
        0x00080000, 0x02000100, 0x40080100, 0x42000100, 0x42080000, 0x00080100,
        0x40000000, 0x02000000, 0x40080000, 0x40080000, 0x00000000, 0x40000100,
        0x42080100, 0x42080100, 0x02000100, 0x42080000, 0x40000100, 0x00000000,
        0x42000000, 0x02080100, 0x02000000, 0x42000000, 0x00080100, 0x00080000,
        0x42000100, 0x00000100, 0x02000000, 0x40000000, 0x02080000, 0x42000100,
        0x40080100, 0x02000100, 0x40000000, 0x42080000, 0x02080100, 0x40080100,
        0x00000100, 0x02000000, 0x42080000, 0x42080100, 0x00080100, 0x42000000,
        0x42080100, 0x02080000, 0x00000000, 0x40080000, 0x42000000, 0x00080100,
        0x02000100, 0x40000100, 0x00080000, 0x00000000, 0x40080000, 0x02080100,
        0x40000100 };
    private static int[] SP6 = { 0x20000010, 0x20400000, 0x00004000,
        0x20404010, 0x20400000, 0x00000010, 0x20404010, 0x00400000, 0x20004000,
        0x00404010, 0x00400000, 0x20000010, 0x00400010, 0x20004000, 0x20000000,
        0x00004010, 0x00000000, 0x00400010, 0x20004010, 0x00004000, 0x00404000,
        0x20004010, 0x00000010, 0x20400010, 0x20400010, 0x00000000, 0x00404010,
        0x20404000, 0x00004010, 0x00404000, 0x20404000, 0x20000000, 0x20004000,
        0x00000010, 0x20400010, 0x00404000, 0x20404010, 0x00400000, 0x00004010,
        0x20000010, 0x00400000, 0x20004000, 0x20000000, 0x00004010, 0x20000010,
        0x20404010, 0x00404000, 0x20400000, 0x00404010, 0x20404000, 0x00000000,
        0x20400010, 0x00000010, 0x00004000, 0x20400000, 0x00404010, 0x00004000,
        0x00400010, 0x20004010, 0x00000000, 0x20404000, 0x20000000, 0x00400010,
        0x20004010 };
    private static int[] SP7 = { 0x00200000, 0x04200002, 0x04000802,
        0x00000000, 0x00000800, 0x04000802, 0x00200802, 0x04200800, 0x04200802,
        0x00200000, 0x00000000, 0x04000002, 0x00000002, 0x04000000, 0x04200002,
        0x00000802, 0x04000800, 0x00200802, 0x00200002, 0x04000800, 0x04000002,
        0x04200000, 0x04200800, 0x00200002, 0x04200000, 0x00000800, 0x00000802,
        0x04200802, 0x00200800, 0x00000002, 0x04000000, 0x00200800, 0x04000000,
        0x00200800, 0x00200000, 0x04000802, 0x04000802, 0x04200002, 0x04200002,
        0x00000002, 0x00200002, 0x04000000, 0x04000800, 0x00200000, 0x04200800,
        0x00000802, 0x00200802, 0x04200800, 0x00000802, 0x04000002, 0x04200802,
        0x04200000, 0x00200800, 0x00000000, 0x00000002, 0x04200802, 0x00000000,
        0x00200802, 0x04200000, 0x00000800, 0x04000002, 0x04000800, 0x00000800,
        0x00200002 };
    private static int[] SP8 = { 0x10001040, 0x00001000, 0x00040000,
        0x10041040, 0x10000000, 0x10001040, 0x00000040, 0x10000000, 0x00040040,
        0x10040000, 0x10041040, 0x00041000, 0x10041000, 0x00041040, 0x00001000,
        0x00000040, 0x10040000, 0x10000040, 0x10001000, 0x00001040, 0x00041000,
        0x00040040, 0x10040040, 0x10041000, 0x00001040, 0x00000000, 0x00000000,
        0x10040040, 0x10000040, 0x10001000, 0x00041040, 0x00040000, 0x00041040,
        0x00040000, 0x10041000, 0x00001000, 0x00000040, 0x10040040, 0x00001000,
        0x00041040, 0x10001000, 0x00000040, 0x10000040, 0x10040000, 0x10040040,
        0x10000000, 0x00040000, 0x10001040, 0x00000000, 0x10041040, 0x00040040,
        0x10000040, 0x10040000, 0x10001000, 0x10001040, 0x00000000, 0x10041040,
        0x00041000, 0x00041000, 0x00001040, 0x00001040, 0x00040040, 0x10000000,
        0x10041000 };

    /* =====================以下为内部使用方法=================== */
    private static void squashBytesToInts(byte[] inBytes, int inOff,
        int[] outInts, int outOff, int intLen)
    {

        for (int i = 0; i < intLen; ++i)
            outInts[outOff + i] = ((inBytes[inOff + i * 4] & 0xff) << 24)
                | ((inBytes[inOff + i * 4 + 1] & 0xff) << 16)
                | ((inBytes[inOff + i * 4 + 2] & 0xff) << 8)
                | (inBytes[inOff + i * 4 + 3] & 0xff);
    }

    private static void spreadIntsToBytes(int[] inInts, int inOff,
        byte[] outBytes, int outOff, int intLen)
    {
        for (int i = 0; i < intLen; ++i)
        {
            outBytes[outOff + i * 4] = (byte) (inInts[inOff + i] >>> 24);
            outBytes[outOff + i * 4 + 1] = (byte) (inInts[inOff + i] >>> 16);
            outBytes[outOff + i * 4 + 2] = (byte) (inInts[inOff + i] >>> 8);
            outBytes[outOff + i * 4 + 3] = (byte) inInts[inOff + i];
        }
    }
    
    private static void makeSMBKey(byte[] key7, byte[] key8)
    {
        key8[0] = (byte) ((key7[0] >> 1) & 0xff);
        key8[1] = (byte) ((((key7[0] & 0x01) << 6) | (((key7[1] & 0xff) >> 2) & 0xff)) & 0xff);
        key8[2] = (byte) ((((key7[1] & 0x03) << 5) | (((key7[2] & 0xff) >> 3) & 0xff)) & 0xff);
        key8[3] = (byte) ((((key7[2] & 0x07) << 4) | (((key7[3] & 0xff) >> 4) & 0xff)) & 0xff);
        key8[4] = (byte) ((((key7[3] & 0x0F) << 3) | (((key7[4] & 0xff) >> 5) & 0xff)) & 0xff);
        key8[5] = (byte) ((((key7[4] & 0x1F) << 2) | (((key7[5] & 0xff) >> 6) & 0xff)) & 0xff);
        key8[6] = (byte) ((((key7[5] & 0x3F) << 1) | (((key7[6] & 0xff) >> 7) & 0xff)) & 0xff);
        key8[7] = (byte) (key7[6] & 0x7F);
        
        for (int i = 0; i < 8; i++)
        {
            key8[i] = (byte) (key8[i] << 1);
        }
    }

}

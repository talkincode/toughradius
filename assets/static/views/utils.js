Date.prototype.format = function(fmt) {
     var o = {
        "M+" : this.getMonth()+1,                 //月份
        "d+" : this.getDate(),                    //日
        "h+" : this.getHours(),                   //小时
        "m+" : this.getMinutes(),                 //分
        "s+" : this.getSeconds(),                 //秒
        "q+" : Math.floor((this.getMonth()+3)/3), //季度
        "S"  : this.getMilliseconds()             //毫秒
    };
    if(/(y+)/.test(fmt)) {
            fmt=fmt.replace(RegExp.$1, (this.getFullYear()+"").substr(4 - RegExp.$1.length));
    }
     for(let k in o) {
        if(new RegExp("("+ k +")").test(fmt)){
             fmt = fmt.replace(RegExp.$1, (RegExp.$1.length===1) ? (o[k]) : (("00"+ o[k]).substr((""+ o[k]).length)));
         }
     }
    return fmt;
}

function hasPerms(session, perms) {
    if (session.level === 'super') {
        return true;
    }
    let flag = false;
    for (let i in perms) {
        if (session.perm_names.indexOf(perms[i]) !== -1) {
            flag = true;
            break;
        }
    }
    return flag;
}

function isManager(session) {
    return session.level === 'super' || session.level === 'master';

}

var format_date = function (date, fmt) {
    var o = {
        "y+": date.getFullYear(),
        "M+": date.getMonth() + 1,                 //月份
        "d+": date.getDate(),                    //日
        "h+": date.getHours(),                   //小时
        "m+": date.getMinutes(),                 //分
        "s+": date.getSeconds(),                 //秒
        "q+": Math.floor((date.getMonth() + 3) / 3), //季度
        "S+": date.getMilliseconds()             //毫秒
    };
    for (var k in o) {
        if (new RegExp("(" + k + ")").test(fmt)) {
            if (k === "y+") {
                fmt = fmt.replace(RegExp.$1, ("" + o[k]).substr(4 - RegExp.$1.length));
            } else if (k === "S+") {
                var lens = RegExp.$1.length;
                lens = lens === 1 ? 3 : lens;
                fmt = fmt.replace(RegExp.$1, ("00" + o[k]).substr(("" + o[k]).length - 1, lens));
            } else {
                fmt = fmt.replace(RegExp.$1, (RegExp.$1.length === 1) ? (o[k]) : (("00" + o[k]).substr(("" + o[k]).length)));
            }
        }
    }
    return fmt;
};


function bytesToSize(_bytes) {
    if(!_bytes||_bytes==="0"){
        return 0
    }
    let bytes = Number(_bytes)
    let sizes = ['bytes', 'K', 'M', 'G', 'T'];
    if (bytes === 0) return '0 Byte';
    let i = Number(Math.floor(Math.log(bytes) / Math.log(1024)));
    return Math.round(bytes / Math.pow(1024, i), 2) + ' ' + sizes[i];
}


function bpsToSize(_bps) {
    let bps = Number(_bps)
    let sizes = ['bps', 'Kbps', 'Mbps', 'Gbps', 'Tbps'];
    if (bps === 0) return '0 bps';
    let i = Number(Math.floor(Math.log(bps) / Math.log(1000)));
    return Math.round(bps / Math.pow(1000, i), 2) + ' ' + sizes[i];
}

function doPost(action, formValues) {
    let form = document.createElement("FORM");
    document.body.appendChild(form);
    form.method = "post";
    form.action = action;
    form.style.display = "none";
    for (let k in formValues) {
        if (k !== 'submit') {
            let _input = document.createElement("input");
            _input.name = k;
            _input.value = formValues[k];
            form.appendChild(_input);
        }
    }
    form.submit();
}

/*
* 判断url是否合法
* */
function checkURL(url) {
    let str = url;
    //判断URL地址的正则表达式为:http(s)?://([\w-]+\.)+[\w-]+(/[\w- ./?%&=]*)?
    //下面的代码中应用了转义字符"\"输出一个字符"/"
    let Expression = /http(s)?:\/\/([\w-]+\.)+[\w-]+(\/[\w- .\/?%&=]*)?/;
    let objExp = new RegExp(Expression);
    return objExp.test(str) === true;
}

function isMobileDevice() {
    let sUserAgent = navigator.userAgent.toLowerCase();
    let bIsIpad = sUserAgent.match(/ipad/i) === 'ipad';
    let bIsIphone = sUserAgent.match(/iphone os/i) === 'iphone os';
    let bIsMidp = sUserAgent.match(/midp/i) === 'midp';
    let bIsUc7 = sUserAgent.match(/rv:1.2.3.4/i) === 'rv:1.2.3.4';
    let bIsUc = sUserAgent.match(/ucweb/i) === 'web';
    let bIsCE = sUserAgent.match(/windows ce/i) === 'windows ce';
    let bIsWM = sUserAgent.match(/windows mobile/i) === 'windows mobile';
    let bIsAndroid = sUserAgent.match(/android/i) === 'android';
    let pathname = location.pathname
    return bIsIpad || bIsIphone || bIsMidp || bIsUc7 || bIsUc || bIsCE || bIsWM || bIsAndroid;

}


function concat2(a, b) {
    return a.concat(b.filter(function (v) {
        return !(a.indexOf(v) > -1)
    }))
}


/**
 * 秒转换友好的显示格式
 * 输出格式：21小时前
 * @param  {[type]} time [description]
 * @return {string}      [description]
 */
function second2Str(time) {
    //存储转换值
    let s;
    if ((time < 60 * 60)) {
        //超过十分钟少于1小时
        s = Math.floor(time / 60);
        return s + "分钟";
    } else if ((time < 60 * 60 * 24) && (time >= 60 * 60)) {
        //超过1小时少于24小时
        s = Math.floor(time / 60 / 60);
        return s + "小时";
    } else if (time >= 60 * 60 * 24) {
        //超过1天少于3天内
        s = Math.floor(time / 60 / 60 / 24);
        return s + "天";
    }
}


function date2str(date, y) {
    let f = function (n) {
        if (n < 10) {
            return "0" + n;
        } else {
            return n;
        }
    };
    let z = {
        y: date.getFullYear(),
        M: f(date.getMonth() + 1),
        d: f(date.getDate()),
        h: f(date.getHours()),
        m: f(date.getMinutes()),
        s: f(date.getSeconds()),
    };
    return y.replace(/(y+|M+|d+|h+|m+|s+)/g, function (v) {
        return ((v.length > 1 ? "0" : "") + eval('z.' + v.slice(-1))).slice(-(v.length > 2 ? v.length : 2))
    });

}

function toDecimal2(x) {
    let f = parseFloat(x);
    if (isNaN(f)) {
        return false;
    }
    f = Math.round(x * 100) / 100;
    let s = f.toString();
    let rs = s.indexOf('.');
    if (rs < 0) {
        rs = s.length;
        s += '.';
    }
    while (s.length <= rs + 2) {
        s += '0';
    }
    return s;
}

function fenToYuan(fen) {
    let num = fen;
    num = fen * 0.01;
    num += '';
    let reg = num.indexOf('.') > -1 ? /(\d{1,3})(?=(?:\d{3})+\.)/g : /(\d{1,3})(?=(?:\d{3})+$)/g;
    num = num.replace(reg, '$1');
    num = toDecimal2(num);
    return num
}

function yuanToFen(yuan, digit) {
    let m = 0,
        s1 = yuan.toString(),
        s2 = digit.toString();
    try {
        m += s1.split(".")[1].length
    } catch (e) {
    }
    try {
        m += s2.split(".")[1].length
    } catch (e) {
    }
    return Number(s1.replace(".", "")) * Number(s2.replace(".", "")) / Math.pow(10, m)
}

function checkTwoPointNum(inputNumber) {
    let partten = /^-?\d+\.?\d{0,2}$/;
    return partten.test(inputNumber)
}

function removeFromArray(arr, key) {
    for (let i = 0; i < arr.length; i++) {
        if (arr[i] === key) {
            arr.splice(i, 1);
        }
    }
}


function syntaxHighlight(json) {
    if (typeof json != 'string') {
         json = JSON.stringify(json, undefined, 2);
    }
    json = json.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
    return json.replace(/("(\\u[a-zA-Z0-9]{4}|\\[^u]|[^\\"])*"(\s*:)?|\b(true|false|null)\b|-?\d+(?:\.\d*)?(?:[eE][+\-]?\d+)?)/g, function (match) {
        let cls = 'number';
        if (/^"/.test(match)) {
            if (/:$/.test(match)) {
                cls = 'key';
            } else {
                cls = 'string';
            }
        } else if (/true|false/.test(match)) {
            cls = 'boolean';
        } else if (/null/.test(match)) {
            cls = 'null';
        }
        return '<span class="' + cls + '">' + match + '</span>';
    });
}


function warpString(src, len) {
    if (!src) {
        return ""
    }
    if (src.length < len) {
        return src
    }
    if (src.length > len) {
        return src.substring(0, len) + "..."
    }
}


function warpRemark(obj) {
    return warpString(obj.remark, 16)
}
var format_date = function (date,fmt) {
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
        if (new RegExp("(" + k + ")").test(fmt)){
            if(k === "y+"){
                fmt = fmt.replace(RegExp.$1, ("" + o[k]).substr(4 - RegExp.$1.length));
            }
            else if(k==="S+"){
                var lens = RegExp.$1.length;
                lens = lens===1?3:lens;
                fmt = fmt.replace(RegExp.$1, ("00" + o[k]).substr(("" + o[k]).length - 1,lens));
            }
            else{
                fmt = fmt.replace(RegExp.$1, (RegExp.$1.length === 1) ? (o[k]) : (("00" + o[k]).substr(("" + o[k]).length)));
            }
        }
    }
    return fmt;
};

function bytesToSize(bytes) {
    var sizes = ['bytes', 'K', 'M', 'G', 'T'];
    if (bytes == 0) return '0 Byte';
    var i = Number(Math.floor(Math.log(bytes) / Math.log(1024)));
    return Math.round(bytes / Math.pow(1024, i), 2) + ' ' + sizes[i];
}

function doPost(action,formValues){
    var form=document.createElement("FORM");
    document.body.appendChild(form);
    form.method = "post";
    form.action = action;
    form.style.display = "none";
    for(var k in formValues){
        if(k!=='submit'){
            var _input = document.createElement("input");
            _input.name=k;
            _input.value=formValues[k];
            form.appendChild(_input);
        }
    }
    form.submit();
}

/*
* 判断url是否合法
* */
function checkURL(url){
    var str = url;
    //判断URL地址的正则表达式为:http(s)?://([\w-]+\.)+[\w-]+(/[\w- ./?%&=]*)?
    //下面的代码中应用了转义字符"\"输出一个字符"/"
    var Expression = /http(s)?:\/\/([\w-]+\.)+[\w-]+(\/[\w- .\/?%&=]*)?/;
    var objExp = new RegExp(Expression);
    if(objExp.test(str) == true){
        return true;
    }else{
        return false;
    }
}

function isMobileDevice(){
    var sUserAgent = navigator.userAgent.toLowerCase();
    var bIsIpad = sUserAgent.match(/ipad/i) == 'ipad';
    var bIsIphone = sUserAgent.match(/iphone os/i) == 'iphone os';
    var bIsMidp = sUserAgent.match(/midp/i) == 'midp';
    var bIsUc7 = sUserAgent.match(/rv:1.2.3.4/i) == 'rv:1.2.3.4';
    var bIsUc = sUserAgent.match(/ucweb/i) == 'web';
    var bIsCE = sUserAgent.match(/windows ce/i) == 'windows ce';
    var bIsWM = sUserAgent.match(/windows mobile/i) == 'windows mobile';
    var bIsAndroid = sUserAgent.match(/android/i) == 'android';
    var pathname = location.pathname
    if(bIsIpad || bIsIphone || bIsMidp || bIsUc7 || bIsUc || bIsCE || bIsWM || bIsAndroid ){
        return true;
    }
    return false;
}


function lc(srcRes) {
    var lang = (navigator.language || navigator.browserLanguage).toLowerCase();
    if ((lang.indexOf('zh') > -1)) {
        // console.log(cnData[srcRes])
        return cnData[srcRes] ? cnData[srcRes] : srcRes;
    } else {
        // console.log(enData[srcRes])
        return enData[srcRes] ? enData[srcRes] : srcRes;
    }
}





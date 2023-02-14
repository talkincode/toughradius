if (!window.wxui)
    window.wxui = {};


if (webix.env.isIE || webix.env.isIE8) {
    webix.message({
        type: "error",
        text: tr("global", "The kernel of the browser you are using is IE, in order to better use the system functions, it is recommended that you use a browser such as chrome or Firefox.")
    });
}

if (!webix.env.touch && webix.ui.scrollSize) {
    webix.CustomScroll.init();
}

webix.editors.$popup = {
    text: {
        view: "popup", width: 320, height: 240,
        body: {view: "textarea"}
    },
    color: {
        view: "popup",
        body: {view: "colorboard", width: 500, height: 500, rows: 50, cols: 50}
    },
    date: {
        view: "popup",
        body: {view: "calendar", weekNumber: true}
    }
};


webix.attachEvent("onAjaxError", function (xhr) {
        let result = /.*?message=(.*),\sinternal.*/.exec(xhr.responseText)
        if (result) {
            webix.message({type: "error", text: result[1], expire: 5000});
        } else {
            webix.message({type: "error", text: xhr.responseText, expire: 5000});
        }

    }
);

webix.attachEvent("onLoadError", function (xhr, view) {
        console.log(xhr.responseText);
        console.log(view.config)
    }
);


// webix.ready(function () {
//     webix.ajax().get("/admin/translate/data").then(function (resp) {
//        window.GlobalTrans = resp.json()
//     })
// })


tr = function (m, v) {
    if (window.GlobalTrans && window.GlobalTrans[m] && window.GlobalTrans[m][v]) {
        return window.GlobalTrans[m][v];
    } else {
        webix.ajax().post("/admin/translate/patch", {
            module: m,
            key: v,
            value: v,
        }).then(function (resp) {
            console.log(resp.msg)
        })
    }
    return v
}

wxui.tr = tr

gtr = function (s) {
    return tr("global", s)
}

wxui.removeFromArray = function (arr, key) {
    for (var i = 0; i < arr.length; i++) {
        if (arr[i] === key) {
            arr.splice(i, 1);
        }
    }
}

wxui.metricsColors = ["#f44336", "#9c27b0", "#3f51b5", "#0288d1", "#009688", "#558b2f", "#ffa000", "#ff5722", "#795548", "#546e7a", "#e91e63", "#1e88e5"]

/**
 * 构造一个工具栏
 * @param config
 * icon
 * title
 * elements
 * @returns {{paddingX: number, view: string, css: string, cols: *[]}}
 */
wxui.getPageToolbar = function (config) {
    return {
        id: config.winid || webix.uid().toString(),
        view: "toolbar",
        paddingX: 10,
        css: "page-toolbar",
        hidden: config.hidden || false,
        cols: [
            {
                view: "label",
                label: " <i class='mdi mdi-" + config.icon + "'></i> " + config.title,
                css: "dash-title-b",
                width: 240,
                align: "left"
            }, {},
            {
                cols: config.elements
            },
        ]
    }
};

/**
 * 构造ICON按钮
 * @param name
 * @param width
 * @param icon
 * @param clickfunc
 * @param config
 * @returns {{view: string, css: string, width: *, icon: string, label: *, type: string, click: *}}
 */
wxui.getIconButton = function (name, width, icon, hidden, clickfunc) {
    return {
        view: "button",
        type: "icon",
        css: "webix_transparent",
        inputWidth: width,
        width: width,
        icon: "mdi mdi-" + icon,
        label: name,
        hidden: hidden,
        click: clickfunc
    };
};

wxui.getIconButton2 = function (name, width, height, icon, hidden, clickfunc) {
    return {
        view: "button",
        type: "icon",
        css: "webix_transparent",
        inputWidth: width,
        width: width,
        height: height,
        icon: "mdi mdi-" + icon,
        label: name,
        hidden: hidden,
        click: clickfunc
    };
};


/**
 * 构造通用按钮
 * @param name
 * @param width
 * @param clickfunc
 * @param config
 * @returns {{view: string, css: string, width: *, label: *, click: *}}
 */
wxui.getPrimaryButton = function (name, width, hidden, clickfunc) {
    return {
        view: "button", css: "webix_primary webix_button", inputWidth: width, width: width, label: name, hidden: hidden,
        click: clickfunc
    };
};

/**
 * 构造删除按钮
 * @param name
 * @param width
 * @param clickfunc
 * @param config
 * @returns {{view: string, css: string, width: *, label: *, click: *}}
 */
wxui.getDangerButton = function (name, width, hidden, clickfunc) {
    return {
        view: "button", css: "webix_danger", inputWidth: width, width: width, label: name, hidden: hidden,
        click: clickfunc
    };
};


/**
 * 获取表格选中的行
 * @param tableid
 * @returns {[]}
 */
wxui.getTableCheckedIds = function (tableid) {
    let rows = [];
    $$(tableid).eachRow(
        function (row) {
            let item = $$(tableid).getItem(row);
            if (item && item.state === 1) {
                rows.push(item.id)
            }
        }
    );
    return rows;
};
wxui.getTableCheckedItems = function (tableid) {
    let items = [];
    $$(tableid).eachRow(
        function (row) {
            let item = $$(tableid).getItem(row);
            if (item && item.state === 1) {
                items.push(item)
            }
        }
    );
    return items;
};

wxui.getTableCheckedAttrs = function (tableid, name) {
    let rows = [];
    $$(tableid).eachRow(
        function (row) {
            let item = $$(tableid).getItem(row);
            if (item && item.state === 1) {
                rows.push(item[name])
            }
        }
    );
    return rows;
};

wxui.getTableChecked_Ids = function (tableid) {
    let rows = [];
    $$(tableid).eachRow(
        function (row) {
            let item = $$(tableid).getItem(row);
            if (item && item.state === 1) {
                rows.push(item._id)
            }
        }
    );
    return rows;
};


/**
 * 构造一个具有标签和关键字的搜索表单
 * @param queryid
 * @param callback
 * @returns {{paddingX: number, elementsConfig: {labelWidth: number}, css: string, view: string, hidden: boolean, elements: {cols: *[]}[], id: *, paddingY: number}}
 */
wxui.getTableQueryStdForm = function (queryid, callback) {
    return {
        id: queryid,
        css: "main-panel-box",
        view: "form",
        hidden: false,
        paddingX: 0,
        paddingY: 0,
        elementsConfig: {labelWidth: 60},
        elements: [
            {
                cols: [
                    {
                        view: "multicombo",
                        name: "tags",
                        label: "",
                        value: "",
                        labelWidth: 0,
                        placeholder: tr("global", "Tags match"),
                        options: "/bss/tag/options",
                        maxWidth: 320,
                    },
                    {
                        view: "text",
                        name: "keyword",
                        label: "",
                        labelWidth: 0,
                        placeholder: tr("global", "Keyword match"),
                        maxWidth: 320,
                    },
                    {
                        cols: [
                            {
                                view: "button", css: "webix_transparent", label: tr("global", "Query"), type: "icon",
                                icon: "mdi mdi-search-web", borderless: true, width: 66, click: function () {
                                    callback();
                                },
                            },
                            {
                                view: "button", css: "webix_transparent", label: tr("global", "Reset"), type: "icon",
                                icon: "mdi mdi-refresh", borderless: true, width: 66, click: function () {
                                    $$(queryid).setValues({keyword: "", tags: ""});
                                },
                            },
                        ],
                    }, {}
                ],

            },
        ],
    };
};


/**
 * 构造一个有关键字的搜索表单
 * @param queryid
 * @param callback
 * @returns {{paddingX: number, elementsConfig: {labelWidth: number}, css: string, view: string, hidden: boolean, elements: {cols: *[]}[], id: *, paddingY: number}}
 */
wxui.getTableQueryKeywordForm = function (queryid, callback) {
    return {
        id: queryid,
        css: "main-panel-box",
        view: "form",
        hidden: false,
        paddingX: 0,
        paddingY: 0,
        elementsConfig: {labelWidth: 60},
        elements: [
            {
                cols: [
                    {
                        view: "text",
                        name: "keyword",
                        label: "",
                        labelWidth: 0,
                        placeholder: tr("global", "Keyword match"),
                        maxWidth: 320,
                    },
                    {
                        cols: [
                            {
                                view: "button", css: "webix_transparent", label: tr("global", "Query"), type: "icon",
                                icon: "mdi mdi-search-web", borderless: true, width: 66, click: function () {
                                    callback();
                                },
                            },
                            {
                                view: "button", css: "webix_transparent", label: tr("global", "Reset"), type: "icon",
                                icon: "mdi mdi-refresh", borderless: true, width: 66, click: function () {
                                    $$(queryid).setValues({keyword: ""});
                                },
                            },
                        ],
                    }, {}
                ],

            },
        ],
    };
};


/**
 * 构造一个有关键字和时间范围的的搜索表单
 * @param queryid
 * @param callback
 * @returns {{paddingX: number, elementsConfig: {labelWidth: number}, css: string, view: string, hidden: boolean, elements: {cols: *[]}[], id: *, paddingY: number}}
 */
wxui.getTableQueryKeyDateRangeForm = function (queryid, callback) {
    return {
        id: queryid,
        css: "main-panel-box",
        view: "form",
        hidden: false,
        paddingX: 10,
        paddingY: 0,
        elementsConfig: {labelWidth: 60},
        elements: [
            {
                cols: [
                    {
                        view: "daterangepicker",
                        name: "date_range",
                        label: tr("global", "Date range"),
                        format: "%Y-%m-%d",
                        width: 300,
                        labelWidth: 80,
                        value: {start: webix.Date.add(new Date(), -1, "day"), end: new Date()}
                    },
                    {
                        view: "text",
                        name: "keyword",
                        label: "",
                        labelWidth: 0,
                        placeholder: tr("global", "Keyword match"),
                        maxWidth: 240,
                    },
                    {
                        cols: [
                            {
                                view: "button", css: "webix_transparent", label: tr("global", "Query"), type: "icon",
                                icon: "mdi mdi-search-web", borderless: true, width: 66, click: function () {
                                    callback();
                                },
                            },
                            {
                                view: "button", css: "webix_transparent", label: tr("global", "Reset"), type: "icon",
                                icon: "mdi mdi-refresh", borderless: true, width: 66, click: function () {
                                    $$(queryid).setValues({keyword: ""});
                                },
                            },
                        ],
                    }, {}
                ],

            },
        ],
    };
};


/**
 * 构造一个具有标签和关键字的搜索表单
 * @param queryid
 * @param elements
 * @returns {{paddingX: number, elementsConfig: {labelWidth: number}, css: string, view: string, hidden: boolean, elements: {cols: *[]}[], id: *, paddingY: number}}
 */
wxui.getTableQueryCustomForm = function (queryid, elements) {
    return {
        id: queryid,
        css: "main-panel-box",
        view: "form",
        hidden: false,
        paddingX: 7,
        paddingY: 0,
        elementsConfig: {labelWidth: 60},
        elements: elements,
    };
};


/**
 * reload 函数
 * @param tableid
 * @param url
 * @param queryid
 * @returns {function(...[*]=)}
 */
wxui.reloadDataFunc = function (tableid, url, queryid) {
    if (!url) {
        url = $$(tableid).config.url;
    }
    if (queryid) {
        return function () {
            $$(tableid).refresh();
            $$(tableid).clearAll();
            let params = $$(queryid).getValues();
            let args = [];
            for (let k in params) {
                let val = params[k];
                if (val instanceof Object) {
                    args.push(k + "=" + webix.stringify(val));
                } else {
                    args.push(k + "=" + val);
                }
            }
            $$(tableid).load(url + '?' + args.join("&"));
        }
    }
    return function () {
        $$(tableid).clearAll();
        $$(tableid).load(url);
    }
};

/**
 * 构造一个数据表
 * @param config
 * tableid
 * columns
 * on
 * url
 * data
 * autoConfig
 * type
 */
wxui.getDatatable = function (config) {
    let uidata = {
        id: config.tableid,
        css: "main-panel-box",
        view: "datatable",
        type: config.type || {},
        headermenu: true,
        select: true,
        resizeColumn: true,
        autoWidth: true,
        autoHeight: true,
        on: config.on || {},
        onClick: config.onClick || {},
    };

    if (config.leftSplit !== undefined) {
        uidata.leftSplit = config.leftSplit
    }
    if (config.datafetch !== undefined) {
        uidata.datafetch = config.datafetch
    }

    if (config.subrow !== undefined) {
        uidata.subrow = config.subrow
    }
    if (config.subview !== undefined) {
        uidata.subview = config.subview
    }

    if (config.loadahead !== undefined) {
        uidata.loadahead = config.loadahead
    }

    if (config.fixedRowHeight !== undefined) {
        uidata.fixedRowHeight = config.fixedRowHeight
    }
    if (config.rowLineHeight !== undefined) {
        uidata.rowLineHeight = config.rowLineHeight
    }
    if (config.scrollX !== undefined) {
        uidata.scrollX = config.scrollX
    }


    if (config.editable !== undefined) {
        uidata.editable = config.editable
        uidata.editaction = "dblclick"
    } else {
        uidata.editable = false
    }

    if (config.rightSplit !== undefined) {
        uidata.rightSplit = config.rightSplit
    }

    if (config.autoConfig) {
        uidata.autoConfig = config.autoConfig
    } else {
        uidata.columns = config.columns
    }

    if (config.pager) {
        uidata.pager = config.tableid + ".dataPager"
    }

    if (config.url) {
        uidata.url = config.url
    }

    if (config.save) {
        uidata.save = config.save
    }

    if (config.data) {
        uidata.data = config.data
    }


    uidata.on["onBeforeLoad"] = function () {
        this.showOverlay("lodding...");
    };

    uidata.on["onAfterLoad"] = function () {
        if (!this.count())
            this.showOverlay(tr("global", "No data loaded"));
        else
            this.hideOverlay();
    };
    return uidata
};

/**
 * 构造一个数据表
 * @param config
 * tableid
 * columns
 * on
 * url
 * data
 */
wxui.getTreetable = function (config) {
    let uidata = {
        id: config.tableid,
        css: "main-panel-box",
        view: "treetable",
        type: config.type || {},
        headermenu: true,
        select: true,
        resizeColumn: true,
        autoWidth: true,
        autoHeight: true,
        on: config.on || {},
        onClick: config.onClick,
    };

    if (config.autoConfig) {
        uidata.autoConfig = config.autoConfig
    } else {
        uidata.columns = config.columns
    }

    if (config.pager) {
        uidata.pager = config.tableid + ".dataPager"
    }
    if (config.url) {
        uidata.url = config.url
    }
    if (config.data) {
        uidata.data = config.data
    }
    uidata.on["onBeforeLoad"] = function () {
        this.showOverlay(tr("global", "loading..."));
    };

    uidata.on["onAfterLoad"] = function () {
        if (!this.count())
            this.showOverlay(tr("global", "No data loaded"));
        else
            this.hideOverlay();
    };
    return uidata
};


/**
 * 构造一个增删改查表格
 * @param config
 * tableid
 * columns
 * query
 * save
 * savefunc
 * @returns {{paddingX: number, elementsConfig: {labelPosition: string}, view: string, elements: *[], scroll: boolean, id: *, paddingY: number}}
 */
wxui.getCurdDatatable = function (config) {
    let tableid = config.tableid || webix.uid();
    let table = {
        id: tableid,
        view: "datatable",
        borderless: true,
        columns: config.columns,
        editable: true,
        select: true,
    };
    if (config.query) {
        table.url = config.query
    }
    if (config.save) {
        table.save = config.save
    }

    let actions = [
        {
            view: "button", css: "webix_transparent", type: "icon", icon: "mdi mdi-refresh",
            label: tr("global", "Refresh"), width: 70, height: 36, click: function () {
                $$(tableid).editStop();
                $$(tableid).clearAll();
                $$(tableid).load(config.query);
            },
        },
        {
            view: "button", css: "webix_transparent", type: "icon", icon: "mdi mdi-plus",
            label: tr("global", "Create"), width: 70, height: 36, click: function () {
                $$(tableid).add(config.inititem || {});
            },
        },
        {
            view: "button", css: "webix_transparent", type: "icon", icon: "mdi mdi-delete",
            label: tr("global", "Delete"), width: 70, height: 36, click: function () {
                $$(tableid).editStop();
                let rid = $$(tableid).getSelectedId();
                $$(tableid).remove(rid);

            }
        }, {}
    ];


    if (config.actions) {
        for (let i in config.actions) {
            actions.push(config.actions[i])
        }
    }

    if (config.savefunc) {
        actions.push({
            view: "button", css: "webix_transparent", type: "icon", icon: "mdi mdi-refresh",
            label: tr("global", "Save"), width: 70, height: 36, click: function () {
                let rows = [];
                $$(tableid).eachRow(
                    function (row) {
                        let item = $$(tableid).getItem(row);
                        rows.push(item)
                    }
                );
                config.savefunc(rows)
            },
        })
    }


    return {
        view: "form",
        scroll: true,
        paddingX: 5,
        paddingY: 5,
        elementsConfig: {
            labelPosition: "left",
        },
        elements: [
            {
                cols: actions,
            },
            {
                cols: [table]
            },
        ],
    }
};

/**
 * 构造一个分页工具栏
 * @param tableid
 * @param callback
 * @returns {{cols: *[], paddingY: number}}
 */
wxui.getTablePagerBar = function (tableid, callback) {
    return {
        paddingY: 3,
        cols: [
            {
                view: "richselect",
                name: "page_num",
                label: tr("global", "Pagesize"),
                value: 20,
                width: 140,
                labelWidth: 45,
                options: [{id: 20, value: "20"},
                    {id: 50, value: "50"},
                    {id: 100, value: "100"},
                    {id: 500, value: "500"},
                    {id: 1000, value: "1000"},
                    {id: 5000, value: "5000"}],
                on: {
                    onChange: function (newv, oldv) {
                        $$(tableid + ".dataPager").define("size", parseInt(newv));
                        $$(tableid + ".dataPager").refresh();
                        if (callback)
                            callback();
                    },
                },
            },
            {
                id: tableid + ".dataPager", view: 'pager', master: false, size: 20, group: 5,
                template: '{common.first()} {common.prev()} {common.pages()} {common.next()} {common.last()} total:#count#',
            }, {},
        ],
    }
};


/**
 * 打开一个表单窗口
 * @param config :
    * winid  窗口ID
 * title  窗口标题
 * width  窗口宽度
 * height 窗口高度
 * elements 表单UI元素
 * post 表单提交地址
 * data 表单初始化数据
 */
wxui.openFormWindow = function (config) {
    let winid = config.winid || webix.uid();
    if ($$(winid))
        return;
    let title = config.title;
    let formid = winid + "_form";
    return webix.ui({
        id: winid,
        view: "window", css: "win-body",
        move: true,
        resize: true,
        modal: true,
        width: config.width || 480,
        height: config.height || 640,
        fullscreen: config.fullscreen || false,
        position: "center",
        head: {
            view: "toolbar",
            css: "win-toolbar",
            cols: [
                {view: "icon", icon: "mdi mdi-laptop", css: "alter"},
                {view: "label", label: title},
                {
                    view: "icon", icon: "mdi mdi-fullscreen-exit", css: "alter", click: function () {
                        webix.fullscreen.exit();
                    }
                },
                {
                    view: "icon", icon: "mdi mdi-fullscreen", css: "alter", click: function () {
                        webix.fullscreen.set($$(winid));
                    }
                },
                {
                    view: "icon", icon: "mdi mdi-close", css: "alter", click: function () {
                        webix.fullscreen.exit();
                        $$(winid).close();

                    }
                }
            ]
        },
        body: {
            rows: [
                {
                    id: formid,
                    view: "form",
                    scroll: true,
                    elementsConfig: config.elementsConfig || {labelWidth: 120},
                    elements: config.elements,
                    data: config.data || {}
                },
                {
                    padding: 5,
                    cols: [{},
                        {
                            view: "button",
                            name: "submit",
                            css: "webix_primary",
                            label: tr("global", "Save"),
                            width: 120,
                            height: 36,
                            click: function () {
                                if (!$$(formid).validate()) {
                                    webix.message({
                                        type: "error",
                                        text: tr("global", "Please fill in the valid data."),
                                        expire: 1000
                                    });
                                    return false;
                                }
                                let param = $$(formid).getValues();
                                if (config.callBefore) {
                                    param = config.callBefore(param)
                                }
                                webix.ajax().post(config.post, param).then(function (result) {
                                    let resp = result.json();
                                    webix.message({type: resp.msgtype, text: resp.msg, expire: 5000});
                                    if (resp.code === 0) {
                                        $$(winid).close();
                                        if (config.callback)
                                            config.callback()
                                    }
                                })
                            }
                        },
                        {
                            view: "button",
                            css: "webix_transparent",
                            icon: "mdi mdi-close",
                            width: 120,
                            label: tr("global", "Cancel"),
                            click: function () {
                                $$(winid).close();
                            }
                        }
                    ]
                }
            ]
        }
    })
};


wxui.openWindow = function (config) {
    let winid = config.winid || webix.uid();
    if ($$(winid))
        return;
    let title = config.title;
    return webix.ui({
        id: winid,
        view: "window", css: "win-body",
        move: true,
        resize: true,
        fullscreen: config.fullscreen || false,
        modal: true,
        width: config.width || 480,
        height: config.height || 640,
        position: "center",
        head: {
            view: "toolbar",
            css: "win-toolbar",
            cols: [
                {view: "icon", icon: "mdi mdi-laptop", css: "alter"},
                {view: "label", label: title},
                {
                    view: "icon", icon: "mdi mdi-fullscreen-exit", css: "alter", click: function () {
                        webix.fullscreen.exit();
                    }
                },
                {
                    view: "icon", icon: "mdi mdi-fullscreen", css: "alter", click: function () {
                        webix.fullscreen.set($$(winid));
                    }
                },
                {
                    view: "icon", icon: "mdi mdi-close", css: "alter", click: function () {
                        webix.fullscreen.exit();
                        if (config.closeEvent) {
                            config.closeEvent()
                        }
                        $$(winid).close();

                    }
                }
            ]
        },
        body: config.body
    })
};


/**
 * 一个表单页面
 * @param config :
    * winid  窗口ID
 * title  窗口标题
 * elements 表单UI元素
 * post 表单提交地址
 * data 表单初始化数据
 * callback
 */
wxui.getForm = function (config) {
    let uid = webix.uid()
    let formid = config.formid || webix.uid();
    let winid = config.winid || webix.uid();
    let pageView = {
        id: winid,
        css: "main-panel",
        padding: 10,
        rows: [
            {
                view: "toolbar",
                css: "page-toolbar",
                cols: [
                    {view: "icon", icon: "mdi mdi-square-edit-outline"},
                    {view: "label", label: config.title},
                    {},
                    config.actions || {
                        cols: [
                            {
                                view: "button",
                                css: "webix_primary",
                                label: tr("global", "Save"),
                                width: 120,
                                height: 36,
                                click: function () {
                                    if (!$$(formid).validate()) {
                                        webix.message({
                                            type: "error",
                                            text: tr("global", "Please fill in the valid data."),
                                            expire: 1000
                                        });
                                        return false;
                                    }
                                    let param = $$(formid).getValues();
                                    webix.ajax().post(config.post, param).then(function (result) {
                                        let resp = result.json();
                                        webix.message({type: resp.msgtype, text: resp.msg, expire: 5000});
                                        if (resp.code === 0) {
                                            $$(bss.core.tabsId).removeOption(winid);
                                            wxui.removeFromArray(bss.core.tabsIds, winid)
                                            if (config.parentid) {
                                                $$(bss.core.tabsId).setValue(config.parentid);
                                            }
                                            if (config.callback)
                                                config.callback(resp)
                                        }

                                    })
                                }
                            },
                            {
                                view: "icon", icon: "mdi mdi-fullscreen", css: "alter", click: function () {
                                    webix.fullscreen.set($$(winid));
                                }
                            },
                        ]
                    }
                ]
            },
            {
                id: formid,
                view: "form",
                css: "detail-form",
                scroll: true,
                elementsConfig: config.elementsConfig || {
                    labelWidth: 120,
                    // inputWidth: 640,
                    bottomPadding: config.bottomPadding || 0,
                },
                elements: config.elements,
                data: config.data,
            }
        ]
    }
    return pageView
};

/**
 * 打开一个侧边栏窗口
 * @param config
 * winid  窗口ID
 * title  窗口标题
 * width  窗口宽度
 * body   窗口UI元素
 */
wxui.openSideWindow = function (config) {
    let sideid = config.winid || webix.uid().toString()
    if ($$(sideid) && $$(sideid).isVisible()) {
        $$(sideid).close();
        return;
    }
    webix.ui({
        view: "sidemenu",
        id: sideid,
        width: config.width || 420,
        position: "right",
        // animate:false,
        // state: function (state) {
        //     let toolbarHeight = $$(config.parentId).$height;
        //     state.top = toolbarHeight;
        //     state.height -= toolbarHeight;
        // },
        body: {
            rows: [
                {
                    css: "panel-toolbar",
                    paddingX: 10,
                    cols: [
                        {view: "icon", icon: "mdi mdi-information-outline", css: "alter webix_transparent"},
                        {view: "label", label: config.title, css: "dash-title-b", inputWidth: 150, align: "left"},
                        {},
                        {
                            view: "icon", icon: "mdi mdi-close", css: "webix_transparent", click: function () {
                                $$(sideid).close();
                            },
                        },
                    ],
                },
                {
                    view: "scrollview",
                    body: config.body
                }
            ],
        },
    }).show();
};

/**
 * 初始化文件上传
 * @param uid
 * @param uploadurl
 * @param callback
 */
wxui.initUploadApi = function (uid, uploadurl, callback) {
    webix.ui({
        id: uid,
        view: "uploader",
        upload: uploadurl,
        on: {
            onBeforeFileAdd: function (item) {
                item.formData = {};
                webix.message({type: "info", text: tr("global", "Upload..."), expire: 3000})
            },
            onFileUpload: function (item) {
                if (callback) {
                    callback(item);
                }
            },
            onFileUploadError: function (item, response) {
                webix.message({type: "error", text: tr("global", "Upload failure"), expire: 3000});
            },
            onUploadComplete: function (resp) {
                webix.message({type: resp.msgtype, text: resp.msg, expire: 5000});
            }
        },
        apiOnly: true
    });
};


/**
 * Segmented 风格的tab页面
 * @param config
 * @returns
 */
wxui.getSegmentedView = function (config) {
    let tabs = {
        view: "segmented",
        id: config.tabsId || webix.uid(),
        value: config.tabsCid,
        multiview: true,
        animate: false,
        align: "left",
        optionWidth: config.tabsWidth,
        padding: 5,
        options: config.tabCells
    }
    let views = {
        animate: false,
        cells: config.viewCells
    }
    return {
        rows: [
            tabs,
            views,
        ]
    }
}


/**
 * 图片选择器
 * @param id
 * @param name
 * @param width
 * @param height
 * @param options
 * @returns
 */
wxui.getPictureSelect = function (id, label, width, height, options) {
    let picview = {
        name: id,
        view: "richselect",
        label: label,
        labelPosition: "top",
        options: {
            view: "datasuggest",
            template: function (obj) {
                return "<img src='" + obj.id + "' style='width:90%;background:gray;'>";
            },
            body: {
                template: function (obj) {
                    return obj.value + "<img style='width:95%;background:gray;padding: 0;' src='" + obj.id + "'>";
                },
                type: {
                    width: 136, height: 136
                },
                xCount: 6,
                scroll: "y",
                autoheight: false,
                height: 480,
                url: options
            }
        }
    }
    if (width && width > 0)
        picview.width = width

    if (height && height > 0)
        picview.height = height

    return picview;
}


/**
 * 导入js文件
 * @param jsname
 * @param session
 * @param callback
 */
wxui.requirejs = function (jsname, session, callback) {
    console.log("load admin/" + jsname + ".js");
    if (session.dev_mode === 'enabled') {
        webix.require("views/" + jsname + ".min.js?ver=" + session.pagever, function () {
            callback();
        });
    } else {
        webix.require("views/" + jsname + ".min.js", function () {
            callback();
        });
    }
};


/**
 * 显示busy进度条
 * @param viewid
 * @param delay
 * @param callback
 */
wxui.showBusyBar = function (viewid, delay, callback) {
    // $$(viewid).disable();
    $$(viewid).showProgress({
        type: "top",
        delay: delay,
        hide: true
    });
    setTimeout(function () {
        callback();
        // $$(viewid).enable();
    }, delay);
};

wxui.getFieldSet = function (name, rows) {
    return {
        view: "fieldset", label: name, body: {
            rows: rows
        }
    }
}

wxui.getWidgetItem = function (item, params) {
    let argstr = ""
    let args = [];
    for (let k in params) {
        let val = params[k];
        if (val instanceof Object) {
            args.push(k + "=" + webix.stringify(val));
        } else {
            args.push(k + "=" + val);
        }
    }
    if (args.length > 0) {
        argstr = '?' + args.join("&")
    }

    let body = {}
    if (item.type === "metrics.html") {
        body = {
            view: "template",
            css: {"background": item.bgcolor + "!important"},
            src: item.src + argstr,
            borderless: true
        };
    }
    if (item.type === "xmetrics") {
        body = {
            view: "template",
            css: {"background": item.bgcolor + "!important"},
            src: item.src + argstr,
            borderless: true
        };
    } else if (item.type === "template") {
        body = {
            rows: [
                {type: "header", template: item.title},
                {view: "template", css: "widget-template", src: item.src + argstr, autoheight: true}
            ]
        };
    } else if (item.type === "iframe") {
        body = {
            rows: [
                {type: "header", template: item.title},
                {view: "iframe", css: "widget-iframe", src: item.src + argstr, autoheight: true}
            ]
        };
    }
    return body;
}


wxui.exportData = function (url, filename) {
    webix.ajax().response("blob").get(url, function (text, data) {
        let a = document.createElement('a');
        let url = window.URL.createObjectURL(data);
        a.href = url;
        a.download = filename;
        a.click();
    });
};

wxui.tryCall = function (callback, errback) {
    try {
        callback()
    } catch (e) {
        console.log(e);
        if (errback) {
            errback(e)
        }
    }
}

wxui.rules = {
    isEmail: function (value) {
        if (!value || value === "") {
            return true
        }
        return (/\S+@[^@\s]+\.[^@\s]+$/).test((value || "").toString());
    },
    isNumber: function (value) {
        return (parseFloat(value) == value);
    },
    isChecked: function (value) {
        return (!!value) || value === "0";
    },
    isNotEmpty: function (value) {
        return (value === 0 || value);
    },
    isMoney: function (value) {
        if (value != null && value !== "") {
            let exp = /^(([1-9]\d*)|\d)(\.\d{1,2})?$/;
            if (!exp.test(value)) {
                return false;
            }
        } else {
            return false;
        }
        return true;
    }
};


wxui.getTableFooterBar = function (config) {
    return {
        paddingY: 3,
        cols: [
            wxui.getIconButton(tr("global", ""), 35, "refresh", false, config.callback),
            {
                view: "richselect",
                name: "page_num",
                label: tr("global", "Size"),
                value: 20,
                width: 150,
                labelWidth: 60,
                options: [{id: 20, value: "" + (config.size || 20)},
                    {id: 50, value: "50"},
                    {id: 100, value: "100"},
                    {id: 500, value: "500"},
                    {id: 1000, value: "1000"},
                    {id: 5000, value: "5000"}],
                on: {
                    onChange: function (newv, oldv) {
                        $$(config.tableid + ".dataPager").define("size", parseInt(newv));
                        $$(config.tableid + ".dataPager").refresh();
                        if (config.callback) {
                            config.callback();
                        }
                    },
                },
            },
            {
                id: config.tableid + ".dataPager", view: 'pager', master: false, size: config.size || 20, group: 5,
                template: '{common.first()} {common.prev()} {common.pages()} {common.next()} {common.last()} total:#count#',
            },
            {},
            {
                hidden: config.actions === undefined || config.actions.length === 0,
                cols: config.actions || []
            }
        ],
    }
};

function warpRemark(obj) {
    return warpString(obj.remark, 16)
}


wxui.openItemDetail = function (config, node) {
    let body = {
        view: "form", paddingX: 20, scroll: "auto",
        elementsConfig: {marginY: 0, labelWidth: 120}, css: "detail-form",
        data: config.data,
        elements: [],
    }

    for (let k in config.data) {
        body.elements.push({view: "text", name: k, label: k, readonly: true})
    }

    webix.ui({
        view: "popup",
        css: "win-body",
        height: 480,
        width: 420,
        scroll: "auto",
        body: {
            rows: [
                {
                    view: "toolbar",
                    css: "win-toolbar",
                    cols: [
                        {view: "icon", icon: "mdi mdi-laptop", css: "alter"},
                        {view: "label", label: config.title},
                    ]
                },
                body
            ]
        },
    }).show(node, {pos: "right"})
}

wxui.confirmCall = function (flag, msg, funccall) {
    if (flag) {
        webix.confirm({
            title: "操作确认", ok: "Yes", cancel: "No", text: msg || "确认此操作吗?", callback: function (ev) {
                if (ev) {
                    funccall()
                }
            }
        });
    } else {
        funccall()
    }
}

wxui.displayTableMessage = function (node, width, height, message, mode) {
    webix.ui({
        view: "popup", height: height, width: width, scroll: "auto", body: {
            name: "content", view: "codemirror-editor", mode: mode, value: message
        }
    }).show(node)
}
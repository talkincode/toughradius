webix.protoUI({
    name: "metrics-group",
    defaults: {
        css: "metrics-group",
        select: false,
        type: {
            height: 75,
            width: "auto",
            type: "tiles",
            template: "<div class='widget-metrics'>" +
                "<div class='icon'><i class='#icon#'></i></div>" +
                "<div class='detail'>" +
                "<div class='value'>#value#</div>" +
                "<div class='name'>#name#</div>" +
                "</div>" +
                "</div>"
        },

    }

}, webix.ui.dataview);


/**
 * echarts
 */
webix.protoUI({
    name: "echarts",
    $init: function () {
        this.uid = "chart" + webix.uid();
        this._chart = null;
        this._waitContent = webix.promise.defer();
        this.$view.innerHTML = "<div id='" + this.uid + "' style='width:100%;height:100%'></div>";
        this.attachEvent("onAfterLoad", function () {
            this._data_ready()
        });
        this.$ready.push(this._render_once);
    },
    _render_once: function () {
        let source = ["/static/echarts/echarts.min.js", "/static/echarts/dark.min.js"];
        webix.require([source])
            .then(webix.bind(this._initChart, this))
            .catch(function (e) {
                this._waitContent.reject(e);
            });
    },
    _initChart: function () {
        let theme = this.config.theme || "light";
        this._chart = echarts.init(this.$view.firstChild, theme);
        this._chart.setOption(this.config.settings);
        this._waitContent.resolve(this._chart)
    },
    _data_ready: function () {
        webix.promise.all([
            this.waitData,
            this._waitContent
        ]).then(webix.bind(this.renderData, this));
    },
    renderData: function () {
        this._chart.setOption(this.data);
    },
    $setSize: function (x, y) {
        if (webix.ui.view.prototype.$setSize.call(this, x, y)) {
            if (this._chart) {
                this._chart.resize();
            }
        }
    },
    getChart: function (wait) {
        return wait ? this._waitContent : this._chart;
    }
}, webix.AtomDataLoader, webix.EventSystem, webix.ui.view);


/**
 * webconsole
 */
webix.protoUI({
    name: "webconsole",
    append: function (logtxt) {
        var old = this.config.template();
        if (old) {
            this.setHTML(old + logtxt);
        } else {
            this.setHTML(logtxt);
        }
    },
    clear: function () {
        this.setHTML("");
    },
    defaults: {
        css: "web-console",
        borderless: true,
        scroll: "y",
    }
}, webix.ui.template);


/**
 * mytemplate
 */
webix.protoUI({
    name: "mytemplate",
    reload: function () {
        let t = this
        webix.ajax().get(this.config.src, {v: new Date().getTime()}).then(function (result) {
            t.setHTML(result.text());
            t.refresh()
        })
    },
    load: function (url) {
        let t = this
        webix.ajax().get(url, {v: new Date().getTime()}).then(function (result) {
            t.setHTML(result.text());
            t.refresh()
        })
    },
    defaults: {
        css: "mymetrics",
        borderless: true,
    }
}, webix.ui.template);



(function() {
    const Torrent = Backbone.Model.extend({
        defaults: function() {
            return {
                title: "",
                torr: {},
                url:"/torrents"
            };
        },
        remove: function() {
            this.destroy();
        },
        fetch: function (){
            const collection = this;
            getTorrent(this.torr.hash, function (torr){
                console.log(torr);
                collection.reset(torr);
            })
        }
    });

    var TorrentList = Backbone.Collection.extend({
        model: Torrent,
        update: function(){
            listTorrent(function(torrs){
                // torrs.forEach(tr=>
                //
                // )
                Torrents.create({title:""});
                console.log(Torrents);
            },function (error) {
                console.log(error);
            });
        }
    });

    var AppView = Backbone.View.extend({
        el: $("#torrents"),
        initialize: function() {
            Torrents.update();
        },
    });

    var Torrents = new TorrentList;
    var App = new AppView;
})();
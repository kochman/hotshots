<template>
    <!--START CARD-->
    <transition mode="out-in">
        <div v-if="startCard" class="animated fadeInUp card">
            <div class="card-header">
                <ul class="nav nav-tabs card-header-tabs pull-right" id="myTab" role="tablist">
                    <li class="nav-item">
                        <a class="nav-link active" id="welcome-tab" data-toggle="tab" href="#welcome" role="tab" aria-controls="welcome" aria-selected="true">Welcome</a>
                    </li>
                    <li class="nav-item">
                        <a class="nav-link" id="help-tab" data-toggle="tab" href="#help" role="tab" aria-controls="help" aria-selected="false">Help</a>
                    </li>
                </ul>
            </div>
            <div class="card-body">
                <div class="tab-content" id="myTabContent">
                    <div class="text-center tab-pane fade show active" id="welcome" role="tabpanel" aria-labelledby="welcome-tab">
                        <h1>🔥</h1>
                        <h1 class="card-title">Hi, welcome to Hotshots!</h1>
                        <h6 class="card-text">Hotshots allows you to curate photos from remote cameras.</h6>
                        <p class="card-text">Before you begin, ensure your pusher is correctly connected to the Hotshots server.</p>
                        <a v-on:click="startCard=false" class="text-white btn btn-primary">Get started</a>
                    </div>
                    <div class="text-center tab-pane fade" id="help" role="tabpanel" aria-labelledby="help-tab">
                        <h1>Pretend there's some helpful information here!</h1>
                        <p>There will be at some point...</p>
                    </div>
                </div>
            </div>
        </div>
    
        <div v-else-if="!startCard && isLoaded === true" class="jumbotron">
            <h1 class="display-4">Your Photos</h1>
            <p class="lead">Photos will automatically appear as they're taken and processed.</p>
            <hr class="my-4">
            <div class="container">
                <div class="row">
                    <div v-for="photoID in photoIDs" class="col-sm">
                        <a v-bind:href="'photos/' + photoID + '/image.jpg'"><img v-bind:src="'photos/' + photoID + '/thumb.jpg'" class="img-thumbnail"></a>
                    </div>
                </div>
            </div>
        </div>
    </transition>
</template>

<script>
    export default {
        data() {
            return {
                title: "Hotshots",
                startCard: true,
                isLoaded: false,
                photoIDs: []
            };
        },
        
        mounted: function() {
            this.fetchAPIData();
        },
        
        methods: {
            fetchAPIData: function() {
                var vue = this;
                var interval=1000; // 1000 = 1 second, 3000 = 3 seconds
                function doAjax() {
                    vue.$http.get('photos/ids').then(function(response) {
                        vue.photoIDs = response.data.ids;
                        vue.isLoaded = true;
                        setTimeout(doAjax, interval);
                    })
                }
                
                setTimeout(doAjax, interval);

            },
            
            setPhotos: function(photos) {
                console.log("TEST");
            }
        }
        
    }
</script>

<template>
  <!--START CARD-->
  <div class="container">
    <h1 class="display-4 my-3">Your Photos</h1>
    <p class="lead">Photos will automatically appear as they're taken and processed.</p>
    <hr class="my-4">
    <div class="container">
      <div v-if="!is_loaded" class="row">
        <div class="mt-5 col-sm text-center spinner"></div>
      </div>
      <div v-else class="row">
        <div v-for="photo_id in photo_ids" class="col-sm">
          <photo :photo_id="photo_id" />
        </div>
      </div>
    </div>
  </div>
</template>

<script>
  import Photo from "./photo";

  export default {
    components: {Photo},
    data() {
      return {
        title: "Hotshots",
        is_loaded: false,
        photo_ids: []
      };
    },

    mounted: function () {
      this.fetchAPIData();
    },

    methods: {
      fetchAPIData: function () {
        var vue = this;
        var interval = 1000; // 1000 = 1 second, 3000 = 3 seconds
        function doAjax() {
          vue.$http.get('photos/ids').then(function (response) {
            vue.photo_ids = response.data.ids;
            vue.is_loaded = true;
            setTimeout(doAjax, interval);
          })
        }

        doAjax()
      },
      setPhotos: function (photos) {
        console.log("TEST");
      }
    }

  }
</script>

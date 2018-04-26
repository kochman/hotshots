<template>
  <div>
    <div class="col-4-lg mt-3 align-items-center text-center">
      <p><a v-bind:href="'photos/' + photo_id + '/image.jpg'">
        <img v-bind:src="'photos/' + photo_id + '/thumb.jpg'" class="img-thumbnail">
      </a></p>
      <div class="d-inline-flex text-center">
        <div class="btn-group" role="group" v-bind:aria-label="'Photo ' + photo_id">
          <b-btn v-b-modal="meta_id" variant="primary">Metadata</b-btn>
          <button type="button" v-on:click="add_tag" class="btn btn-secondary">Tag</button>
          <button type="button" v-on:click="delete_photo" class="btn btn-danger">Delete</button>
        </div>
      </div>
    </div>
    <b-modal v-bind:id="meta_id" size="lg" :ok-only=true title="Photo Metadata">
      <table class="table">
        <thead>
        <tr>
          <th scope="col" width="20%">Key</th>
          <th scope="col" width="80%">Value</th>
        </tr>
        </thead>
        <tbody>
          <tr v-for="(value, key) in mdata">
            <td>{{ key }}</td>
            <td>{{ value }}</td>
          </tr>
        </tbody>
      </table>
    </b-modal>
  </div>


</template>

<script>
  import bootbox from "bootbox"
  import bModal from 'bootstrap-vue/es/components/modal/modal'
  import bBtn from 'bootstrap-vue/es/components/button/button'

  export default {
    name: "photo",
    components: {
      "b-modal": bModal,
      "b-btn": bBtn
    },
    data(){
      return {
        mdata: {}
      }
    },
    computed: {
      meta_id: function () {
        return "meta-" + this.photo_id
      }
    },
    props: [
      "photo_id",
    ],
    methods: {
      delete_photo: function () {
        bootbox.confirm("Do you want to delete this photo?", function (confirmed) {
          if (confirmed) {
            this.$http.delete('photos/' + this.photo_id).then(function (response) {
              if (!response.data.success) {
                bootbox.alert("Unable to delete photo.");
              } else {
                location.reload(true);
              }
            }.bind(this), function (response) {
              bootbox.alert("Unable to delete photo.");
            }.bind(this))
          }
        }.bind(this))
      },
      get_metadata: function() {
        this.$http.get('photos/' + this.photo_id + "/meta").then(function (response) {
          this.mdata = response.data.photo
        }.bind(this));
      },
      add_tag: function() {
        bootbox.prompt("Enter tag", function (tag) {
          if(tag !== null){
            this.$http.post('photos/' + this.photo_id + "/tags/" + tag).then(
              function (response) {
                this.get_metadata()
              }.bind(this), function (response) {
                bootbox.alert("Unable to add tag: " + tag);
              }.bind(this)
            )
          }
        }.bind(this))
      }
    },
    mounted: function () {
      this.get_metadata()
    }
  }

</script>

<style scoped>

</style>

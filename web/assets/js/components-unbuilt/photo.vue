<template>
  <div>
    <div class="col-4-lg mt-3 align-items-center text-center">
      <p><a v-bind:href="'photos/' + photo_id + '/image.jpg'">
        <img v-bind:src="'photos/' + photo_id + '/thumb.jpg'" class="img-thumbnail">
      </a></p>
      <div class="d-inline-flex text-center">
        <div class="btn-group" role="group" v-bind:aria-label="'Photo ' + photo_id">
          <b-btn @click="show_metadata=true" variant="primary">Metadata</b-btn>
          <b-btn @click="show_addtag=true" variant="secondary">Add Tag</b-btn>
          <button type="button" v-on:click="delete_photo" class="btn btn-danger">Delete</button>
        </div>
      </div>
    </div>
    <b-modal v-model="show_metadata" size="lg" :ok-only=true title="Photo Metadata">
      <table class="table mx-0 my-0 py-0 px-0">
        <thead>
        <tr>
          <th scope="col" width="20%">Key</th>
          <th scope="col" width="80%">Value</th>
        </tr>
        </thead>
        <tbody>
          <tr v-for="(value, key) in mdata">
            <td>{{ key }}</td>
            <td v-if="key==='tags'" class="my-0">
              <ul>
                <li class="py-1" v-for="tag in value">{{tag}}&nbsp;<b-btn size="sm" variant="danger" @click="delete_tag(tag)">X</b-btn></li>
              </ul>
            </td>
            <td v-else>{{value}}</td>
          </tr>
        </tbody>
      </table>
      <div slot="modal-footer" class="w-100">
        <b-btn class="float-right" variant="primary" @click="show_metadata=false">
          Close
        </b-btn>
      </div>
    </b-modal>
    <b-modal v-model="show_addtag" size="lg" :ok-only=true title="Enter Tag" :hide-footer="true">
      <form>
        <input title="Tag" class="form-control" v-model="tag" v-on:keydown.prevent.enter="add_tag"/>
        <span class="float-right mt-3">
          <b-btn variant="primary" @click="add_tag">Add</b-btn>
          <b-btn variant="danger" @click="show_addtag=false">Cancel</b-btn>
        </span>
      </form>
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
        mdata: {},
        show_metadata: false,
        show_addtag: false,
        tag: "",
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
                this.$emit('deleted')
              }
            }.bind(this), function () {
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
        this.$http.post('photos/' + this.photo_id + "/tags/" + this.tag).then(
          function () {
            this.get_metadata();
            this.show_addtag = false;
            this.tag = ""
          }.bind(this), function () {
            bootbox.alert("Unable to add tag: " + this.tag);
          }.bind(this))
      },
      delete_tag: function(tag) {
        this.$http.delete('photos/' + this.photo_id + "/tags/" + tag).then(
          function () {
            this.get_metadata();
          }.bind(this), function () {
            bootbox.alert("Unable to delete tag: " + tag);
          }.bind(this))
      }
    },
    mounted: function () {
      this.get_metadata()
    },
    watch: {
      photo_id: function (){
        this.get_metadata()
      }
    }
  }

</script>

<style scoped>

</style>

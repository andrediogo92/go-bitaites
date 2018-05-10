
Vue.component('post-component', {
    props: ['timeline'],
    data: function () {
        return {
            message: ""
        }
    },
    template: `
        <div class="card" id="post">
            <h5 class="card-header">Post Box</h5>
            <div class="card-body">
                <textarea v-model="message" placeholder="What's on your mind?"></textarea>
                <button class="btn btn-dark" v-on:click="registerMessage">Post</button>             
            </div>
        </div>
    `,
    methods: {
        registerMessage: function (event) {
            timeline.addPost(message);
        }
    }
});

Vue.component('timeline-component', {
    props: ['timeline'],
    template: `
        <div class="container">
            <ul class="list-group">
                <post-item v-for="post in timeline" v-bind:post="post"></post-item>
            </ul>
        </div>
    `
});

Vue.component('post-item', {
    props: ['post'],
    template: `
        <li class="list-group-item-dark">
            <div class="card">
                <div class="card-header">
                    {{post.user}}
                </div>
                <p class="card-text">{{post.text}}</p>
                <iframe v-if="post.url != null && post.url.length > 0" v-bind:src="post.url"></iframe>
            </div>
        </li>
    `
});


let vm = new Vue({
    el: '#app',
    template: `
        <div class="container">
            <post-component timeline="t"></post-component>
            <timeline-component timeline="t.getPostHistory()"></timeline-component>
        </div>
    `,
    data: {t: timeline}
});

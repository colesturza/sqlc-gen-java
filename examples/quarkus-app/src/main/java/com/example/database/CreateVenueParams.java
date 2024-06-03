// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
//   sqlc-gen-java dev

package com.example.database;

public class CreateVenueParams {

    private final String slug;
    private final String name;
    private final String city;
    private final String spotifyPlaylist;
    private final Status status;
    private final java.util.List<Status> statuses;
    private final java.util.List<String> tags;

    CreateVenueParams(final String slug, final String name, final String city, final String spotifyPlaylist, final Status status, final java.util.List<Status> statuses, final java.util.List<String> tags) {
        this.slug = slug;
        this.name = name;
        this.city = city;
        this.spotifyPlaylist = spotifyPlaylist;
        this.status = status;
        this.statuses = statuses;
        this.tags = tags;
    }

    public String getSlug() {
        return this.slug;
    }

    public String getName() {
        return this.name;
    }

    public String getCity() {
        return this.city;
    }

    public String getSpotifyPlaylist() {
        return this.spotifyPlaylist;
    }

    public Status getStatus() {
        return this.status;
    }

    public java.util.Optional<java.util.List<Status>> getStatuses() {
        return java.util.Optional.ofNullable(this.statuses);
    }

    public java.util.Optional<java.util.List<String>> getTags() {
        return java.util.Optional.ofNullable(this.tags);
    }

    public static BuilderCreateVenueParams builder() {
        return new BuilderCreateVenueParams();
    }

    public static class BuilderCreateVenueParams {

        private String slug;
        private String name;
        private String city;
        private String spotifyPlaylist;
        private Status status;
        private java.util.List<Status> statuses;
        private java.util.List<String> tags;

        BuilderCreateVenueParams() {}

        public BuilderCreateVenueParams slug(final String slug) {
            this.slug = slug;
            return this;
        }

        public BuilderCreateVenueParams name(final String name) {
            this.name = name;
            return this;
        }

        public BuilderCreateVenueParams city(final String city) {
            this.city = city;
            return this;
        }

        public BuilderCreateVenueParams spotifyPlaylist(final String spotifyPlaylist) {
            this.spotifyPlaylist = spotifyPlaylist;
            return this;
        }

        public BuilderCreateVenueParams status(final Status status) {
            this.status = status;
            return this;
        }

        public BuilderCreateVenueParams status(final Status status) {
            if (this.statuses == null) {
                this.statuses = new java.util.ArrayList<Status>();
            }
            this.statuses.add(status);
            return this;
        }

        public BuilderCreateVenueParams statuses(final java.util.Collection<? extends Status> statuses) {
            if (this.statuses == null) {
                this.statuses = new java.util.ArrayList<Status>();
            }
            this.statuses.addAll(statuses);
            return this;
        }

        public BuilderCreateVenueParams clearStatuses() {
            if (this.statuses != null) {
                this.statuses.clear();
            }
            return this;
        }

        public BuilderCreateVenueParams tag(final String tag) {
            if (this.tags == null) {
                this.tags = new java.util.ArrayList<String>();
            }
            this.tags.add(tag);
            return this;
        }

        public BuilderCreateVenueParams tags(final java.util.Collection<? extends String> tags) {
            if (this.tags == null) {
                this.tags = new java.util.ArrayList<String>();
            }
            this.tags.addAll(tags);
            return this;
        }

        public BuilderCreateVenueParams clearTags() {
            if (this.tags != null) {
                this.tags.clear();
            }
            return this;
        }

        public CreateVenueParams build() {
            return new CreateVenueParams(slug, name, city, spotifyPlaylist, status, statuses, tags);
        }
    }
}
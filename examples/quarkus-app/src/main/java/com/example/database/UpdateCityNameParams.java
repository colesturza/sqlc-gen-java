// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
//   sqlc-gen-java dev

package com.example.database;

public class UpdateCityNameParams {

    private final String slug;
    private final String name;

    UpdateCityNameParams(final String slug, final String name) {
        this.slug = slug;
        this.name = name;
    }

    public String getSlug() {
        return this.slug;
    }

    public String getName() {
        return this.name;
    }

    public static BuilderUpdateCityNameParams builder() {
        return new BuilderUpdateCityNameParams();
    }

    public static class BuilderUpdateCityNameParams {

        private String slug;
        private String name;

        BuilderUpdateCityNameParams() {}

        public BuilderUpdateCityNameParams slug(final String slug) {
            this.slug = slug;
            return this;
        }

        public BuilderUpdateCityNameParams name(final String name) {
            this.name = name;
            return this;
        }

        public UpdateCityNameParams build() {
            return new UpdateCityNameParams(slug, name);
        }
    }
}
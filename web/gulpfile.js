/* eslint-disable import/no-extraneous-dependencies */
const gulp = require('gulp')
const inlinesource = require('gulp-inline-source')
const replace = require('gulp-replace')

gulp.task('default', () =>
  gulp
    .src('./build/*.html')
    .pipe(replace('.js"></script>', '.js" inline></script>'))
    .pipe(replace('rel="stylesheet">', 'rel="stylesheet" inline>'))
    .pipe(
      inlinesource({
        compress: false,
        ignore: ['png'],
      }),
    )
    .pipe(gulp.dest('./dest')),
)

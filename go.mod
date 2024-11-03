module github.com/imharish-sivakumar/modern-oauth2-system

replace (
	github.com/hariharan-sivakumar/modern-oauth2-system/service-utils => ./service-utils
	github.com/imharish-sivakumar/modern-oauth2-system/proto => ./cisauth-proto
)

go 1.21

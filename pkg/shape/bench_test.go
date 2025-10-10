package shape

import (
	"testing"

	"github.com/shapestone/shape/internal/parser"
)

// Simple schema benchmarks - single object with few properties
func BenchmarkParse_JSONV_Simple(b *testing.B) {
	input := `{"id": UUID, "name": String(1, 100)}`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Parse(parser.FormatJSONV, input)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParse_PropsV_Simple(b *testing.B) {
	input := `id=UUID
name=String(1, 100)`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Parse(parser.FormatPropsV, input)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParse_XMLV_Simple(b *testing.B) {
	input := `<schema><id>UUID</id><name>String(1, 100)</name></schema>`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Parse(parser.FormatXMLV, input)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParse_CSVV_Simple(b *testing.B) {
	input := `id,name
UUID,String(1, 100)`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Parse(parser.FormatCSVV, input)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParse_YAMLV_Simple(b *testing.B) {
	input := `id: UUID
name: String(1, 100)`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Parse(parser.FormatYAMLV, input)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParse_TEXTV_Simple(b *testing.B) {
	input := `id: UUID
name: String(1, 100)`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Parse(parser.FormatTEXTV, input)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Medium schema benchmarks - nested objects with arrays
func BenchmarkParse_JSONV_Medium(b *testing.B) {
	input := `{
		"user": {
			"id": UUID,
			"username": String(3, 20),
			"email": Email,
			"age": Integer(18, 120),
			"roles": [String(1, 50)]
		},
		"metadata": {
			"created": ISO-8601,
			"tags": [String(1, 30)]
		}
	}`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Parse(parser.FormatJSONV, input)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParse_PropsV_Medium(b *testing.B) {
	input := `user.id=UUID
user.username=String(3, 20)
user.email=Email
user.age=Integer(18, 120)
user.roles[]=String(1, 50)
metadata.created=ISO-8601
metadata.tags[]=String(1, 30)`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Parse(parser.FormatPropsV, input)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParse_XMLV_Medium(b *testing.B) {
	input := `<schema>
		<user>
			<id>UUID</id>
			<username>String(3, 20)</username>
			<email>Email</email>
			<age>Integer(18, 120)</age>
			<roles>String(1, 50)</roles>
		</user>
		<metadata>
			<created>ISO-8601</created>
			<tags>String(1, 30)</tags>
		</metadata>
	</schema>`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Parse(parser.FormatXMLV, input)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParse_CSVV_Medium(b *testing.B) {
	input := `id,username,email,age,created
UUID,String(3,20),Email,Integer(18,120),ISO-8601`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Parse(parser.FormatCSVV, input)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParse_YAMLV_Medium(b *testing.B) {
	input := `user:
  id: UUID
  username: String(3, 20)
  email: Email
  age: Integer(18, 120)
  roles:
    - String(1, 50)
metadata:
  created: ISO-8601
  tags:
    - String(1, 30)`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Parse(parser.FormatYAMLV, input)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParse_TEXTV_Medium(b *testing.B) {
	input := `user.id: UUID
user.username: String(3, 20)
user.email: Email
user.age: Integer(18, 120)
user.roles[]: String(1, 50)
metadata.created: ISO-8601
metadata.tags[]: String(1, 30)`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Parse(parser.FormatTEXTV, input)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Large schema benchmarks - deeply nested with many properties
func BenchmarkParse_JSONV_Large(b *testing.B) {
	input := `{
		"user": {
			"id": UUID,
			"username": String(3, 20),
			"email": Email,
			"age": Integer(18, 120),
			"active": true,
			"profile": {
				"firstName": String(1, 50),
				"lastName": String(1, 50),
				"bio": String(0, 500),
				"avatar": URL,
				"social": {
					"twitter": String(1, 15),
					"github": String(1, 39),
					"linkedin": URL
				}
			},
			"roles": [String(1, 50)],
			"permissions": [String(1, 100)],
			"settings": {
				"theme": String(1, 20),
				"language": String(2, 5),
				"timezone": String(1, 50),
				"notifications": {
					"email": true,
					"push": true,
					"sms": false
				}
			}
		},
		"metadata": {
			"created": ISO-8601,
			"updated": ISO-8601,
			"version": Integer(1, +),
			"tags": [String(1, 30)],
			"categories": [String(1, 50)]
		}
	}`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Parse(parser.FormatJSONV, input)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParse_PropsV_Large(b *testing.B) {
	input := `user.id=UUID
user.username=String(3, 20)
user.email=Email
user.age=Integer(18, 120)
user.active=true
user.profile.firstName=String(1, 50)
user.profile.lastName=String(1, 50)
user.profile.bio=String(0, 500)
user.profile.avatar=URL
user.profile.social.twitter=String(1, 15)
user.profile.social.github=String(1, 39)
user.profile.social.linkedin=URL
user.roles[]=String(1, 50)
user.permissions[]=String(1, 100)
user.settings.theme=String(1, 20)
user.settings.language=String(2, 5)
user.settings.timezone=String(1, 50)
user.settings.notifications.email=true
user.settings.notifications.push=true
user.settings.notifications.sms=false
metadata.created=ISO-8601
metadata.updated=ISO-8601
metadata.version=Integer(1, +)
metadata.tags[]=String(1, 30)
metadata.categories[]=String(1, 50)`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Parse(parser.FormatPropsV, input)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParse_XMLV_Large(b *testing.B) {
	input := `<schema>
		<user>
			<id>UUID</id>
			<username>String(3, 20)</username>
			<email>Email</email>
			<age>Integer(18, 120)</age>
			<active>true</active>
			<profile>
				<firstName>String(1, 50)</firstName>
				<lastName>String(1, 50)</lastName>
				<bio>String(0, 500)</bio>
				<avatar>URL</avatar>
				<social>
					<twitter>String(1, 15)</twitter>
					<github>String(1, 39)</github>
					<linkedin>URL</linkedin>
				</social>
			</profile>
			<roles>String(1, 50)</roles>
			<permissions>String(1, 100)</permissions>
			<settings>
				<theme>String(1, 20)</theme>
				<language>String(2, 5)</language>
				<timezone>String(1, 50)</timezone>
				<notifications>
					<email>true</email>
					<push>true</push>
					<sms>false</sms>
				</notifications>
			</settings>
		</user>
		<metadata>
			<created>ISO-8601</created>
			<updated>ISO-8601</updated>
			<version>Integer(1, +)</version>
			<tags>String(1, 30)</tags>
			<categories>String(1, 50)</categories>
		</metadata>
	</schema>`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Parse(parser.FormatXMLV, input)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParse_CSVV_Large(b *testing.B) {
	input := `id,username,email,age,firstName,lastName,bio,avatar,twitter,github,linkedin,theme,language,timezone,created,updated
UUID,String(3,20),Email,Integer(18,120),String(1,50),String(1,50),String(0,500),URL,String(1,15),String(1,39),URL,String(1,20),String(2,5),String(1,50),ISO-8601,ISO-8601`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Parse(parser.FormatCSVV, input)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParse_YAMLV_Large(b *testing.B) {
	input := `user:
  id: UUID
  username: String(3, 20)
  email: Email
  age: Integer(18, 120)
  active: true
  profile:
    firstName: String(1, 50)
    lastName: String(1, 50)
    bio: String(0, 500)
    avatar: URL
    social:
      twitter: String(1, 15)
      github: String(1, 39)
      linkedin: URL
  roles:
    - String(1, 50)
  permissions:
    - String(1, 100)
  settings:
    theme: String(1, 20)
    language: String(2, 5)
    timezone: String(1, 50)
    notifications:
      email: true
      push: true
      sms: false
metadata:
  created: ISO-8601
  updated: ISO-8601
  version: Integer(1, +)
  tags:
    - String(1, 30)
  categories:
    - String(1, 50)`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Parse(parser.FormatYAMLV, input)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParse_TEXTV_Large(b *testing.B) {
	input := `user.id: UUID
user.username: String(3, 20)
user.email: Email
user.age: Integer(18, 120)
user.active: true
user.profile.firstName: String(1, 50)
user.profile.lastName: String(1, 50)
user.profile.bio: String(0, 500)
user.profile.avatar: URL
user.profile.social.twitter: String(1, 15)
user.profile.social.github: String(1, 39)
user.profile.social.linkedin: URL
user.roles[]: String(1, 50)
user.permissions[]: String(1, 100)
user.settings.theme: String(1, 20)
user.settings.language: String(2, 5)
user.settings.timezone: String(1, 50)
user.settings.notifications.email: true
user.settings.notifications.push: true
user.settings.notifications.sms: false
metadata.created: ISO-8601
metadata.updated: ISO-8601
metadata.version: Integer(1, +)
metadata.tags[]: String(1, 30)
metadata.categories[]: String(1, 50)`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Parse(parser.FormatTEXTV, input)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ParseAuto benchmark - auto-detection overhead
func BenchmarkParseAuto_JSONV(b *testing.B) {
	input := `{"id": UUID, "name": String(1, 100)}`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := ParseAuto(input)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseAuto_YAMLV(b *testing.B) {
	input := `id: UUID
name: String(1, 100)`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := ParseAuto(input)
		if err != nil {
			b.Fatal(err)
		}
	}
}

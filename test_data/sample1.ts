export const environment = {
    name: "{{varfile.v1.name}}",
    age: {{varfile.v1.age}},
    address: {
        street: "{{address.street}}",
        city: "{{address.city}}",
        state: "{{address.state}}",
        zip: "{{address.zip}}"
    },
    skills: {{varfile.v1.skills}},
    education: {
        degree: "{{education.degree}}",
        major: "{{education.major}}",
        university: "{{education.university}}",
        raw: "{{education}}"
    },
    serverPassword: "{{env.secrets.password}}",
    serverVars: {
        email: "{{env.vars.email}}",
    },
    notDefined: "{{notDefined}}",
    notDefinedVar: "{{env.vars.notDefinedVar}}",
    notDefinedSecret: "{{env.secrets.notDefinedSecret}}"
};
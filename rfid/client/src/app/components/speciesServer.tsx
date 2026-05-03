import {Note} from "@/app/components/formSubcomponents/notes";
import {EntryPerms} from "@/app/components/perms";
import {ACL} from "@/app/components/accessControlServer";

export function TestSpeciesOk() {
    const a: SpeciesData = {
        _id: "(ID_HERE)",
        scientificName: "(SCI_NAME_HERE)",
        aliases: ["(Alias 1)", "(Alias 2)"],
        standardSubstrate: "(SUBSTRATE ID)",
        notes: [{
            time: 123,
            note: "(NOTE 1)"
        }, {
            time: 456,
            note: "(NOTE 2)"
        }],
        lastUpdated: 789,
        //perms: {userPerms: {ids:[{id:"userCollId",val:"userName"}],canWrite:[true]},projectPerms: {ids:["proj1","proj2"],canWrite:[true, false]}, blanketPerms: 1},
    }
    return a
}

export interface SpeciesData {
    _id: string
    scientificName: string
    aliases?: string[]
    standardSubstrate: string
    notes?: Note[]
    lastUpdated: number
    acl?: ACL
    defaultAcl?: ACL
}

import {Note} from "@/app/components/formSubcomponents/notes";
import {EntryPerms} from "@/app/components/perms";
import {TestNotes} from "@/app/components/formSubcomponents/contaminations";
import {ACL} from "@/app/components/accessControlServer";

export function TestSubspeciesOk(){
    const a: SubspeciesData = {
        _id: "(SUBSPECIES NAME HERE)",
        species: "(SPECIES NAME HERE)",
        aliases: ["(Alias 1)","(Alias 2)"],
        notes: TestNotes,
        lastUpdated: 789,
        //perms: {userPerms: {ids:[{id:"userCollId",val:"userName"}],canWrite:[true]},projectPerms: {ids:["proj1","proj2"],canWrite:[true, false]}, blanketPerms: 1},
    }
    return a
}

export interface SubspeciesData {
    _id: string
    species: string
    aliases?: string[]
    notes?: Note[]
    lastUpdated: number
    acl?: ACL
    defaultAcl?: ACL
}

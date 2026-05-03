import {Note} from "@/app/components/formSubcomponents/notes";
import {PicWithNotesIncoming} from "@/app/components/formSubcomponents/picWithNotes";
import {Contamination} from "@/app/components/formSubcomponents/contaminations";
import {EntryPerms} from "@/app/components/perms";
import {ACL} from "@/app/components/accessControlServer";

export function TestBagOk(){ // TODO: DELETEME // TODO: FIXME!
    const a: BagData = {
        _id: "(BAG ID HERE)",
        recipe: "(SUB RECIPE)",
        //substrateBatch: // TODO: this
        wetness: 5,
        pcRun: "(PC RUN)",
        filterSize: "(FILTER SIZE)",
        creationDate: Date.now()-2000,
        genSpore: 7,
        genFruitOrSpore:2,
        sealDate: Date.now()-1000,
        knownFruitable: true,
        species: "(SPECIES)",
        subspecies: "(SUBSPECIES)",
        innoc: "(INNOC ID)",
        transfersOut: ["(TRANSFER OUT 1)","(TRANSFER OUT 2)"],
        parentType: "plate",
        parent: "(PARENT ID)",
        pics: [], // TODO: this???
        contamination: [], // TODO: THIS?
        mostRecentImage: undefined, // TODO: ?
        flushes: [], // TODO: ?
        sale: "(SALE_ID)",
        disposed: Date.now()+5000,
        notes: [{time: Date.now(),note: "(TEST NOTE 1)"},{time: Date.now()+2000,note: "(TEST NOTE 2)"}],
        lastUpdated: 789,
        //perms: {userPerms: {ids:[{id:"userCollId",val:"userName"}],canWrite:[true]},projectPerms: {ids:["proj1","proj2"],canWrite:[true, false]}, blanketPerms: 1},
    }
    return a
}

export interface BagData {
    _id: string
    recipe: string // Substrate recipe
    substrateBatch?: string // TODO: do this on box. Add other new fields to boxes
    pcRun?: string
    filterSize: string
    creationDate: number
    genSpore?: number // TODO: NEW
    genFruitOrSpore?: number // TODO: NEW
    sealDate?: number
    wetness?: number // TODO: handle everywhere
    knownFruitable?: boolean
    species?: string
    subspecies?: string
    innoc?: string
    transfersOut?: string[]
    parentType?: string
    parent?: string
    //projects?: string[] // TODO: NEW
    pics?: PicWithNotesIncoming[]
    contamination?: Contamination[]
    mostRecentImage?: PicWithNotesIncoming // TODO: used to be string
    flushes?: PicWithNotesIncoming[]
    sale?: string // TODO: NEW
    disposed?: number
    notes?: Note[]
    lastUpdated: number
    acl?: ACL
}

// TODO: bag selector. RFID or text input selector
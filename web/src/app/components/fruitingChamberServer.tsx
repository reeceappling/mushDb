import {Note} from "@/app/components/formSubcomponents/notes";
import {
    ExamplePicWithNotesIncoming,
    PicWithNotesIncoming
} from "@/app/components/formSubcomponents/picWithNotes";
import {
    Contamination,
    ExampleContaminations,
    ExamplePicsWithNotesIncoming
} from "@/app/components/formSubcomponents/contaminations";
import {ACL, TestAcl} from "@/app/components/accessControlServer";
import CloseableSelector, {SelectorProps} from "@/app/components/selector";
import {FruitingChamberSelector} from "@/app/components/fruitingChamberClient";

// export function TestFruitingChamberOk(){
//     return new FruitingChamberData({
//         _id: "(FC ID HERE)",
//         recipe: "(SUB RECIPE)",
//         substrateBatch: "(SUB BATCH)",
//         cupsGrain: 4,
//         mixedSubstratePerGrain: 1,
//         casingPerGrain: 0.5,
//         creationDate: Date.now()-2000,
//         species: "(SPECIES)",
//         subspecies: "(SUBSPECIES)",
//         innoc: "(INNOC ID)",
//         genSpore: 7,
//         genFruitOrSpore:2,
//         transfersOut: ["(TRANSFER OUT 1)","(TRANSFER OUT 2)"],
//         parentType: "plate",
//         parent: "(PARENT ID)",
//         pics: ExamplePicsWithNotesIncoming,
//         contamination: ExampleContaminations,
//         flushes: ExamplePicsWithNotesIncoming,
//         knownFruitable: true,
//         mostRecentImage: ExamplePicWithNotesIncoming,
//         sale: "(SALE_ID)",
//         disposed: Date.now()+5000,
//         notes: [{time: Date.now(),note: "(TEST NOTE 1)"},{time: Date.now()+2000,note: "(TEST NOTE 2)"}],
//         lastUpdated: 789,
//         acl: TestAcl(),
//     })
// }

export interface FruitingChamberData {
    _id: string
    creationDate: number
    recipe: string // Substrate recipe
    substrateBatch?: string
    cupsGrain: number
    mixedSubstratePerGrain: number
    casingPerGrain: number
    species?: string
    subspecies?: string
    innoc?: string
    genSpore?: number
    genFruitOrSpore?: number
    transfersOut?: string[]
    parentType?: string
    parent?: string
    pics?: PicWithNotesIncoming[]
    contamination?: Contamination[]
    flushes?: PicWithNotesIncoming[]
    knownFruitable?: boolean
    mostRecentImage?: PicWithNotesIncoming
    sale?: string
    disposed?: number
    notes?: Note[]
    lastUpdated: number
    acl: ACL
}
export class FruitingChamberData {
    // Accept a single object containing the fields
    constructor(init?: Partial<FruitingChamberData>) {
        // Dynamically map the object fields onto the class instance
        Object.assign(this, init);
    }

    public getId(): string {
        return this._id
    }
    public entryType(): string {
        return "fruitingChamber"
    }
    public description(): string {
        if(this.species !== undefined){
            const kfSent = this.knownFruitable!==undefined?"":(this.knownFruitable?"Known fruitable.":"Nonfruitable")
            const contamsSent = (this.contamination!==undefined&&this.contamination.length!==0)?`${this.contamination.length} contam notes.`:"Not noted as contaminated."
            const flushesSent = (this.flushes!==undefined&&this.flushes.length!==0)?` ${this.flushes.length} flushes. `:""
            return `Fruiting chamber ${this._id}. Species: ${this.species}. ${this.subspecies!==undefined&&`Subspecies: ${this.subspecies}`}. ${kfSent}. ${contamsSent}${flushesSent}Created on ${new Date(this.creationDate).toISOString()}. Last updated on ${new Date(this.lastUpdated).toISOString()}` // TODO: KF, FilterSize, flushes, contams, disposal date, etc?
        }
        if(this.disposed !== undefined){
            return `Fruiting chamber ${this._id}. Disposed on ${new Date(this.disposed).toISOString()} after ${this.flushes?this.flushes.length:0} flushes`
        }
        return `Fruiting chamber ${this._id}. Not innoculated. Created on ${new Date(this.creationDate).toISOString()}.`
        // TODO: substrate recipe?
    }
}

// export function FruitingChamberSelectorCloseable(sp: SelectorProps<FruitingChamberData>) { // TODO: use
//     const doSel = (val?: FruitingChamberData):void=>{
//         if (!val){
//             return
//         }
//         sp.doSelect(val)
//     }
//     return <CloseableSelector<FruitingChamberData> props={{
//         allowCreation: sp.allowCreation,
//         doSelect: doSel, // For selecting normally
//         closeTxt: "Close Fruiting Chamber List",
//         //createTxt: "Create Bag",// TODO: ???
//         lowercase: "fruiting chamber",
//         //creatorInPage: sp.creatorInPage,// TODO: ???
//         //createEndpt: "bag",// TODO: ???
//         createSelector:(selHdl: (onSelect: FruitingChamberData) => void)=>{
//             return <FruitingChamberSelector allowCreate={sp.allowCreation} doSelect={(v)=>{
//                 v && selHdl(v)
//             }}/>
//         },
//         // TODO: ok?
//         // createCreator:(selHdl: (onSelect: FruitingChamberData) => void)=>{
//         //     return <NewFruitingChamberForm handlers={{onCreate: selHdl, isTopLevel: false}}/>
//         // },
//     }}/>
// }
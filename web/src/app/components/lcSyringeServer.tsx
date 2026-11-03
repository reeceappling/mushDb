import {Note} from "@/app/components/formSubcomponents/notes";
import {ACL} from "@/app/components/accessControlServer";
import CloseableSelector, {SelectorProps} from "@/app/components/selector";
import {LcSyringeSelector} from "@/app/components/lcSyringeClient";
import {PicWithNotesIncoming} from "@/app/components/formSubcomponents/picWithNotes";

// export function TestLcSyringeOk(){
//     return new LcSyringeData({
//         _id: "(LC ID HERE)",
//         creationDate: Date.now()-2000,
//         species: "(SPECIES NAME)",
//         subspecies: "(SUBSPECIES NAME)",
//         confirmedClean: undefined,
//         knownFruitable: true,
//         genSpore: 7,
//         genFruitOrSpore: 3,
//         transfersOut: ["(TRANSFER 1)","(TRANSFER 2)"],
//         parent: "(PARENT ID)",
//         disposed: Date.now()+40000,
//         notes: [{time: Date.now(), note: "(TEST NOTE 1)"}, {time: Date.now() + 2000, note: "(TEST NOTE 2)"}],
//         lastUpdated: 789,
//         acl: TestAcl(),
//     })
// }
export interface LcSyringeData {
    _id: string
    parent?: string
    creationDate: number
    species: string
    subspecies?: string
    sale?: string
    genSpore?:  number
    genFruitOrSpore?: number
    confirmedClean?: boolean
    knownFruitable?: boolean
    transfersOut?: string[]
    disposed?: number
    pics?: PicWithNotesIncoming[]
    mostRecentImage?: PicWithNotesIncoming
    notes?: Note[]
    lastUpdated: number
    acl: ACL
}
export class LcSyringeData {
    // Accept a single object containing the fields
    constructor(init?: Partial<LcSyringeData>) {
        // Dynamically map the object fields onto the class instance
        Object.assign(this, init);
    }

    public getId(): string {
        return this._id
    }
    public entryType(): string {
        return "lcSyringe"
    }
    public description(): string {
        return `Liquid culture syringe ${this._id}. Species ${this.species}. ${this.subspecies!==undefined&&`Subspecies ${this.subspecies}`}. Created on ${new Date(this.creationDate).toISOString()}. Last updated on ${new Date(this.lastUpdated).toISOString()}.${this.disposed!==undefined&&` Disposed on ${new Date(this.disposed).toISOString()}.`}` // TODO: KF, FilterSize, flushes, contams, disposal date, etc? TRANSFERS?
    }
}

export function LcSyringeSelectorCloseable(sp: SelectorProps<LcSyringeData>) { // TODO: use
    const doSel = (val?: LcSyringeData):void=>{
        if (!val){
            return
        }
        sp.doSelect(val)
    }
    return <CloseableSelector<LcSyringeData> props={{
        allowCreation: sp.allowCreation,
        doSelect: doSel, // For selecting normally
        closeTxt: "Close LC Syringe List",
        //createTxt: "Create Bag",// TODO: ???
        lowercase: "liquid culture syringe",
        //creatorInPage: sp.creatorInPage,// TODO: ???
        //createEndpt: "bag",// TODO: ???
        createSelector:(selHdl: (onSelect: LcSyringeData) => void)=>{
            return <LcSyringeSelector doSelect={(v)=>{
                v && selHdl(v)
            }}/>
        },
        // TODO: ok?
        // createCreator:(selHdl: (onSelect: FruitingChamberData) => void)=>{
        //     return <NewFruitingChamberForm handlers={{onCreate: selHdl, isTopLevel: false}}/>
        // },
    }}/>
}
import {Note} from "@/app/components/formSubcomponents/notes";
import {
    Contamination, ExampleContaminations, ExamplePicsWithNotesIncoming,
} from "@/app/components/formSubcomponents/contaminations";
import {
    ExamplePicWithNotesIncoming,
    PicWithNotesIncoming
} from "@/app/components/formSubcomponents/picWithNotes";
import {ACL, TestAcl} from "@/app/components/accessControlServer";
import CloseableSelector, {SelectorProps} from "@/app/components/selector";
import {LcSelector} from "@/app/components/lcClient";

export function TestLcOk(){
    return new LcData({
        _id: "(LC ID HERE)",
        recipe: "(LC RECIPE ID HERE)",
        creationDate: Date.now()-2000,
        species: "(SPECIES NAME)",
        subspecies: "(SUBSPECIES NAME)",
        innoc: "(Innoc transfer id!)",
        genSpore: 7,
        genFruitOrSpore: 3,
        transfersOut: ["(TRANSFER 1)","(TRANSFER 2)"],
        parentType: "lc",
        parent: "(PARENT ID)",
        pics: ExamplePicsWithNotesIncoming,
        confirmedClean: true,
        contamination: ExampleContaminations,
        knownFruitable: true,
        disposed: Date.now()+40000,
        mostRecentImage: ExamplePicWithNotesIncoming,
        notes: [{time: Date.now(), note: "(TEST NOTE 1)"}, {time: Date.now() + 2000, note: "(TEST NOTE 2)"}],
        lastUpdated: 789,
        acl: TestAcl(),
    })
}
export interface LcData {
    _id: string
    recipe: string
    pcRun?: string
    creationDate: number
    species?: string
    subspecies?: string
    innoc?: string
    genSpore?:  number
    genFruitOrSpore?: number
    transfersOut?: string[]
    parentType?: string
    parent?: string
    pics?: PicWithNotesIncoming[]
    confirmedClean?: boolean
    contamination?: Contamination[]
    knownFruitable?: boolean
    disposed?: number
    mostRecentImage?: PicWithNotesIncoming
    notes?: Note[]
    lastUpdated: number
    acl: ACL
}
export class LcData {
    // Accept a single object containing the fields
    constructor(init?: Partial<LcData>) {
        // Dynamically map the object fields onto the class instance
        Object.assign(this, init);
    }

    public getId(): string {
        return this._id
    }
    public entryType(): string {
        return "lc"
    }
    public description(): string {
        if(this.species !== undefined){
            const kfSent = this.knownFruitable!==undefined?"":(this.knownFruitable?"Known fruitable.":"Nonfruitable")
            const contamsSent = (this.contamination!==undefined&&this.contamination.length!==0)?`${this.contamination.length} contam notes.`:"Not noted as contaminated."
            return `Liquid culture ${this._id}. Species: ${this.species}. ${this.subspecies!==undefined&&`Subspecies: ${this.subspecies}`}. ${kfSent}. ${contamsSent} Created on ${new Date(this.creationDate).toISOString()}. Last updated on ${new Date(this.lastUpdated).toISOString()}` // TODO: KF, FilterSize, flushes, contams, disposal date, etc?
        }
        if(this.disposed !== undefined){
            return `Liquid culture ${this._id}. Disposed on ${new Date(this.disposed).toISOString()}.`
        }
        return `Liquid culture ${this._id}. Not innoculated. Created on ${new Date(this.creationDate).toISOString()}.`
    }
}

export function LcSelectorCloseable(sp: SelectorProps<LcData>) { // TODO: use
    const doSel = (val?: LcData):void=>{
        if (!val){
            return
        }
        sp.doSelect(val)
    }
    return <CloseableSelector<LcData> props={{
        allowCreation: sp.allowCreation,
        doSelect: doSel, // For selecting normally
        closeTxt: "Close LC List",
        //createTxt: "Create Bag",// TODO: ???
        lowercase: "liquid culture",
        //creatorInPage: sp.creatorInPage,// TODO: ???
        //createEndpt: "bag",// TODO: ???
        createSelector:(selHdl: (onSelect: LcData) => void)=>{
            return <LcSelector allowCreate={sp.allowCreation} hideDisposed={sp.hideDisposed} doSelect={(v)=>{
                v && selHdl(v)
            }}/>
        },
        // TODO: ok?
        // createCreator:(selHdl: (onSelect: FruitingChamberData) => void)=>{
        //     return <NewFruitingChamberForm handlers={{onCreate: selHdl, isTopLevel: false}}/>
        // },
    }}/>
}
import {Note} from "@/app/components/formSubcomponents/notes";
import {ACL, TestAcl} from "@/app/components/accessControlServer";
import CloseableSelector, {SelectorProps} from "@/app/components/selector";
import {MssSelector, NewMssForm} from "@/app/components/mssClient";
import {PicWithNotesIncoming} from "@/app/components/formSubcomponents/picWithNotes";


export function TestMssOk(){
    return new MssData({
        _id: "(MSS ID HERE)",
        creationDate: Date.now()-2000,
        species: "(SPECIES NAME)",
        subspecies: "(SUBSPECIES NAME)",
        parent: "(PARENT ID)",
        transfersOut: ["(TRANSFER 1)","(TRANSFER 2)"],
        sale: "(SALE ID)",
        disposed: Date.now()+40000,
        notes: [{time: Date.now(),note: "(TEST NOTE 1)"},{time: Date.now()+2000,note: "(TEST NOTE 2)"}],
        lastUpdated: 789,
        acl: TestAcl(),
    })
}
export interface MssData {
    _id: string
    creationDate: number
    species: string
    subspecies?: string
    parent?: string
    transfersOut?: string[]
    sale?: string
    disposed?: number
    pics?: PicWithNotesIncoming[]
    mostRecentImage?: PicWithNotesIncoming
    notes?: Note[]
    lastUpdated: number
    acl: ACL
}
export class MssData {
    // Accept a single object containing the fields
    constructor(init?: Partial<MssData>) {
        // Dynamically map the object fields onto the class instance
        Object.assign(this, init);
    }

    public getId(): string {
        return this._id
    }
    public entryType(): string {
        return "mss"
    }
    public description(): string {
        return `Multispore syringe ${this._id}. Species ${this.species}.${this.subspecies!==undefined&&` Subspecies ${this.subspecies}.`} Created on ${new Date(this.creationDate).toISOString()}. Last updated on ${new Date(this.lastUpdated).toISOString()}.${this.disposed!==undefined&&` Disposed on ${new Date(this.disposed).toISOString()}`}` // TODO: KF
    }
}

export function MssSelectorCloseable(sp: SelectorProps<MssData>) {
    const doSel = (val?: MssData):void=>{
        if (!val){
            return
        }
        sp.doSelect(val)
    }
    return <CloseableSelector<MssData> props={{
        allowCreation: sp.allowCreation,
        doSelect: doSel, // For selecting normally
        closeTxt: "Close MSS List",
        createTxt: "Create MSS",
        lowercase: "mss",
        creatorInPage: sp.creatorInPage,
        createEndpt: "mss",
        createSelector:(selHdl: (onSelect: MssData) => void)=>{
            return <MssSelector allowCreate={sp.allowCreation} doSelect={(v)=>{
                v && selHdl(v)
            }}/>
        },
        createCreator:(selHdl: (onSelect: MssData) => void)=>{
            return <NewMssForm handlers={{onCreate: selHdl, isTopLevel: false}}/>
        },
    }}/>
}
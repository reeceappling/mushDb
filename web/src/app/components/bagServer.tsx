import {Note} from "@/app/components/formSubcomponents/notes";
import {PicWithNotesIncoming} from "@/app/components/formSubcomponents/picWithNotes";
import {Contamination} from "@/app/components/formSubcomponents/contaminations";
import {ACL, TestAcl} from "@/app/components/accessControlServer";
import CloseableSelector, {SelectorProps} from "@/app/components/selector";
import {BagSelector} from "@/app/components/bagClient";

export function TestBagOk(){ // TODO: DELETEME? // TODO: FIXME!
    return new BagData({
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
        acl: TestAcl(),
    })
}

export interface BagData {
    _id: string
    recipe: string // Substrate recipe
    substrateBatch?: string
    pcRun?: string
    filterSize: string
    creationDate: number
    genSpore?: number
    genFruitOrSpore?: number
    sealDate?: number
    wetness?: number
    knownFruitable?: boolean
    species?: string
    subspecies?: string
    innoc?: string
    transfersOut?: string[]
    parentType?: string
    parent?: string
    pics?: PicWithNotesIncoming[]
    contamination?: Contamination[]
    mostRecentImage?: PicWithNotesIncoming
    flushes?: PicWithNotesIncoming[]
    sale?: string
    disposed?: number
    notes?: Note[]
    lastUpdated: number
    acl: ACL
}
export class BagData {
    // Accept a single object containing the fields
    constructor(init?: Partial<BagData>) {
        // Dynamically map the object fields onto the class instance
        Object.assign(this, init);
    }

    public getId(): string {
        return this._id
    }
    public entryType(): string {
        return "bag"
    }
}

export function BagSelectorCloseable(sp: SelectorProps<BagData>) { // TODO: use
    const doSel = (val?: BagData):void=>{
        if (!val){
            return
        }
        sp.doSelect(val)
    }
    return <CloseableSelector<BagData> props={{
        allowCreation: sp.allowCreation,
        doSelect: doSel, // For selecting normally
        closeTxt: "Close Bag List",
        //createTxt: "Create Bag",// TODO: ???
        lowercase: "bag",
        //creatorInPage: sp.creatorInPage,// TODO: ???
        //createEndpt: "bag",// TODO: ???
        createSelector:(selHdl: (onSelect: BagData) => void)=>{
            return <BagSelector allowCreate={sp.allowCreation} doSelect={(v)=>{
                v && selHdl(v)
            }}/>
        },
        // TODO: ok?
        // createCreator:(selHdl: (onSelect: BagData) => void)=>{
        //     return <NewBagForm handlers={{onCreate: selHdl, isTopLevel: false}}/>
        // },
    }}/>
}

// TODO: bag selector. RFID or text input selector